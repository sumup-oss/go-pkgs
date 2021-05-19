// Copyright 2019 SumUp Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rabbitmq

import (
	"context"
	"time"

	"github.com/sumup-oss/go-pkgs/task"

	"github.com/palantir/stacktrace"

	"github.com/streadway/amqp"
)

type RabbitMQClientInterface interface {
	CreateChannel(ctx context.Context) (*amqp.Channel, error)
	Setup(ctx context.Context, setup *Setup) error
	Close() error
}

type ClientConfig struct {
	// ConnectionURI is the string used to connect to rabbitmq, e.g `amqp://...`
	ConnectionURI string
	// Metric is an interface to collect metrics about the client and consumer
	// There is NullMetric struct if you want to skip them
	Metric Metric
	// ConnectRetryAttempts number of attempts to try and dial the rabbitmq, and create a channel
	ConnectRetryAttempts int
	// InitialReconnectDelay delay between each attempt
	InitialReconnectDelay time.Duration
}

// A simple client that tries to connect to rabbitmq and create a channel
//
// Does not attempt to reconnect if the connection drops
type RabbitMQClient struct {
	amqpURI               string
	conn                  *amqp.Connection
	metric                Metric
	connectRetryAttempts  int
	initialReconnectDelay time.Duration
	cfg                   *ClientConfig
}

func NewRabbitMQClient(ctx context.Context, cfg *ClientConfig) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		amqpURI:               cfg.ConnectionURI,
		metric:                cfg.Metric,
		connectRetryAttempts:  cfg.ConnectRetryAttempts,
		initialReconnectDelay: cfg.InitialReconnectDelay,
		cfg:                   cfg,
	}

	err := task.RetryUntil(cfg.ConnectRetryAttempts, cfg.InitialReconnectDelay, func(c context.Context) error {
		conn, dialErr := amqp.Dial(client.amqpURI)
		if dialErr != nil {
			cfg.Metric.ObserveRabbitMQConnectionRetry()
			return task.NewRetryableError(dialErr)
		}

		client.conn = conn
		client.metric.ObserveRabbitMQConnection()
		return nil
	})(ctx)

	if err != nil {
		client.metric.ObserveRabbitMQChanelConnectionFailed()
		return nil, stacktrace.Propagate(err, "couldn't dial rabbitmq")
	}

	return client, nil
}

func (c *RabbitMQClient) CreateChannel(ctx context.Context) (*amqp.Channel, error) {
	var channel *amqp.Channel

	err := task.RetryUntil(c.cfg.ConnectRetryAttempts, c.cfg.InitialReconnectDelay, func(ctx context.Context) error {
		var channelErr error

		channel, channelErr = c.conn.Channel()
		if channelErr != nil {
			c.metric.ObserveRabbitMQChanelConnectionRetry()
			return task.NewRetryableError(channelErr)
		}

		c.metric.ObserveRabbitMQChanelConnection()
		return nil
	})(ctx)

	if err != nil {
		c.metric.ObserveRabbitMQChanelConnectionFailed()
		return nil, stacktrace.Propagate(err, "couldn't create channel for rabbitmq")
	}

	return channel, nil
}

func (c *RabbitMQClient) Setup(ctx context.Context, setup *Setup) error {
	channel, err := c.CreateChannel(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create a RMQ channel")
	}

	for _, e := range setup.Exchanges {
		err := channel.ExchangeDeclare(e.Name, e.Kind, e.Durable, e.AutoDelete, e.Internal, e.NoWait, e.Args)
		if err != nil {
			return stacktrace.Propagate(err, "could not declare exchange")
		}
	}

	for _, q := range setup.Queues {
		_, err := channel.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Args)
		if err != nil {
			return stacktrace.Propagate(err, "could not declare queue")
		}
	}

	for _, b := range setup.QueueBindings {
		err := channel.QueueBind(b.Name, b.Key, b.Exchange, b.NoWait, b.Args)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"could not bind queue %s to exchange %s", b.Name, b.Exchange,
			)
		}
	}

	return nil
}

func (c *RabbitMQClient) Close() error {
	err := c.conn.Close()
	return stacktrace.Propagate(err, "RMQ connection close")
}
