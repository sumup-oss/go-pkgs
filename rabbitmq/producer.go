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
	"errors"
	"sync/atomic"

	"github.com/palantir/stacktrace"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

var ErrProducerConnection = errors.New("RMQ producer has already closed the connection")

// MessageArgs captures the fields related to the message sent to the server.
type MessageArgs struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers amqp.Table

	// Correlation identifier
	CorrelationID string
}

type Producer struct {
	client  RabbitMQClientInterface
	logger  logger.StructuredLogger
	metric  Metric
	channel *amqp.Channel

	closeCh chan *amqp.Error

	// Needs to be thread safe since publish can be called from multiple goroutines.
	// That is why we need atomic here.
	isClosed int32
}

func NewProducer(client RabbitMQClientInterface, logger logger.StructuredLogger, metric Metric) (*Producer, error) {
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
		isClosed: 0,
	}, nil
}

func (p *Producer) Publish(
	exchange,
	key string,
	mandatory,
	immediate bool,
	expiration string,
	body []byte,
	args MessageArgs,
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
				tracingField(args.CorrelationID),
			)
		} else {
			p.logger.Warn(
				"RMQ closed the connection without an error",
				tracingField(args.CorrelationID),
			)
		}
		atomic.CompareAndSwapInt32(&p.isClosed, 0, 1)
	default:
		if atomic.LoadInt32(&p.isClosed) == 1 {
			return stacktrace.Propagate(ErrProducerConnection, "RabbitMQ connection closed")
		}

		err := p.channel.Publish(
			exchange,
			key,
			mandatory,
			immediate,
			amqp.Publishing{
				Headers:       args.Headers,
				CorrelationId: args.CorrelationID,
				Expiration:    expiration,
				Body:          body,
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
