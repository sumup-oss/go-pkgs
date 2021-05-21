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

	"github.com/sumup-oss/go-pkgs/backoff"
	"github.com/sumup-oss/go-pkgs/logger"
	"github.com/sumup-oss/go-pkgs/task"

	"github.com/palantir/stacktrace"

	"github.com/streadway/amqp"
)

type RabbitMQClientInterface interface {
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

	BackoffConfig *backoff.Config
}

// A simple client that tries to connect to rabbitmq and create a channel
//
// Does not attempt to reconnect if the connection drops
type RabbitMQClient struct {
	amqpURI              string
	metric               Metric
	connectRetryAttempts int
	cfg                  *ClientConfig
}

func NewRabbitMQClient(ctx context.Context, cfg *ClientConfig) RabbitMQClientInterface {
	return &RabbitMQClient{
		amqpURI:              cfg.ConnectionURI,
		metric:               cfg.Metric,
		connectRetryAttempts: cfg.ConnectRetryAttempts,
		cfg:                  cfg,
	}
}

func (c *RabbitMQClient) CreateConsumer(
	ctx context.Context,
	handler Handler,
	logger logger.StructuredLogger,
	cfg ConsumerConfig,
) (*Consumer, error) {
	var (
		conn    *amqp.Connection
		channel *amqp.Channel
	)

	err := task.RetryWithBackoff(
		c.cfg.ConnectRetryAttempts,
		backoff.NewBackoff(c.cfg.BackoffConfig),
		func(ctx context.Context) error {
			var dialErr error
			conn, dialErr = amqp.Dial(c.cfg.ConnectionURI)
			if dialErr != nil {
				c.cfg.Metric.ObserveRabbitMQConnectionRetry()
				return task.NewRetryableError(dialErr)
			}

			c.metric.ObserveRabbitMQConnection()

			var chanErr error
			channel, chanErr = c.createChannel(ctx, conn)
			if chanErr != nil {
				return task.NewRetryableError(chanErr)
			}

			return nil
		})(ctx)

	if err != nil {
		c.metric.ObserveRabbitMQChanelConnectionFailed()
		return nil, stacktrace.Propagate(err, "couldn't dial rabbitmq")
	}

	return newConsumer(conn, channel, handler, logger, c.metric, cfg), nil
}

func (c *RabbitMQClient) createChannel(
	ctx context.Context,
	conn *amqp.Connection,
) (*amqp.Channel, error) {

	channel, channelErr := conn.Channel()
	if channelErr != nil {
		c.metric.ObserveRabbitMQChanelConnectionFailed()
		return nil, stacktrace.Propagate(channelErr, "couldn't create channel for rabbitmq")
	}

	c.metric.ObserveRabbitMQChanelConnection()

	return channel, nil
}

func (c *RabbitMQClient) Setup(ctx context.Context, setup *Setup) error {
	// channel, err := c.createChannel(ctx)
	// if err != nil {
	// return stacktrace.Propagate(err, "failed to create a RMQ channel")
	// }

	// for _, e := range setup.Exchanges {
	// err := channel.ExchangeDeclare(e.Name, e.Kind, e.Durable, e.AutoDelete, e.Internal, e.NoWait, e.Args)
	// if err != nil {
	// return stacktrace.Propagate(err, "could not declare exchange")
	// }
	// }

	// for _, q := range setup.Queues {
	// _, err := channel.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Args)
	// if err != nil {
	// return stacktrace.Propagate(err, "could not declare queue")
	// }
	// }

	// for _, b := range setup.QueueBindings {
	// err := channel.QueueBind(b.Name, b.Key, b.Exchange, b.NoWait, b.Args)
	// if err != nil {
	// return stacktrace.Propagate(
	// err,
	// "could not bind queue %s to exchange %s", b.Name, b.Exchange,
	// )
	// }
	// }

	return nil
}

func (c *RabbitMQClient) Close() error {
	return nil
}
