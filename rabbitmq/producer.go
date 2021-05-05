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

	"github.com/palantir/stacktrace"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

type Producer struct {
	client  *RabbitMQClient
	logger  logger.StructuredLogger
	metric  Metric
	channel *amqp.Channel

	closeCh  chan *amqp.Error
	isClosed bool
}

func NewProducer(client *RabbitMQClient, logger logger.StructuredLogger, metric Metric) (*Producer, error) {
	channel, err := client.CreateChannel(context.TODO())
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create a channel")
	}

	return &Producer{
		client:   client,
		logger:   logger,
		metric:   metric,
		channel:  channel,
		closeCh:  channel.NotifyClose(make(chan *amqp.Error)),
		isClosed: false,
	}, nil
}

func (p *Producer) Publish(
	exchange,
	key string,
	mandatory,
	immediate bool,
	expiration string,
	body []byte,
	args amqp.Table,
) error {
	select {
	case rmqErr := <-p.closeCh:
		if rmqErr != nil {
			p.logger.Warn(
				"RMQ closed the connection",
				zap.String("reason", rmqErr.Reason),
				zap.Int("code", rmqErr.Code),
				zap.Bool("recover", rmqErr.Recover),
				zap.Bool("server", rmqErr.Server),
			)
		} else {
			p.logger.Warn(
				"RMQ closed the connection with nil rmqErr",
			)
		}
		p.isClosed = true
	default:
		if p.isClosed {
			return stacktrace.NewError("RMQ producer has already closed the connection")
		}

		err := p.channel.Publish(
			exchange,
			key,
			mandatory,
			immediate,
			amqp.Publishing{
				Headers:    args,
				Expiration: expiration,
				Body:       body,
			},
		)
		p.metric.ObserveMsgPublish(err == nil)

		return stacktrace.Propagate(err, "failed to publish RMQ message")
	}

	return nil
}

func (p *Producer) Close() error {
	err := p.client.Close()
	return stacktrace.Propagate(err, "failed to close RMQ producer")
}
