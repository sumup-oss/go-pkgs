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

	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"

	"github.com/streadway/amqp"
)

// A consumer that is works with Handler interface
// It needs a RabbitMQClient to work with and is started with the Run() method
type RabbitMQConsumer struct {
	client  *RabbitMQClient
	done    chan bool
	handler Handler
	logger  logger.StructuredLogger
	metric  Metric
}

func NewConsumer(
	client *RabbitMQClient,
	handler Handler,
	logger logger.StructuredLogger,
	metric Metric,
) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		client:  client,
		done:    make(chan bool),
		handler: handler,
		logger:  logger,
		metric:  metric,
	}
}

func (c *RabbitMQConsumer) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		c.logger.Info("Received context cancel. Going to close rabbit connections.")
		_ = c.client.channel.Cancel(c.handler.GetConsumerTag(), false)

		if !c.handler.WaitToConsumeInflight() {
			c.client.channel.Close()
		}

		<-c.done
		c.logger.Info("handler stopped")
		_ = c.client.Close()
	}()

	if ctx.Err() != nil {
		return stacktrace.Propagate(ctx.Err(), "context canceled")
	}

	deliveries, err := c.client.channel.Consume(
		c.handler.GetQueueName(),
		c.handler.GetConsumerTag(),
		c.handler.QueueAutoAck(),
		c.handler.ExclusiveConsumer(),
		false,
		false,
		nil,
	)
	if err != nil {
		return stacktrace.Propagate(err, "couldn't start consuming from channel")
	}

	err = c.handleDeliveries(ctx, deliveries)

	return stacktrace.Propagate(err, "stopped consumer")
}

// nolint:gocognit
func (c *RabbitMQConsumer) handleDeliveries(ctx context.Context, deliveries <-chan amqp.Delivery) error {
	for {
		select {
		case d := <-deliveries:
			// nolint:godox
			// TODO: until we add better logging level conditions
			//c.logger.Debug(
			//	"msg delivered",
			//	zap.Uint64("tag", d.DeliveryTag),
			//	zap.ByteString("body", d.Body),
			//)

			c.metric.ObserveMsgDelivered()

			acknowledgement, err := c.handler.ReceiveMessage(ctx, d.Body)
			if err != nil {
				return stacktrace.Propagate(err, "handler returned error")
			}

			if c.handler.QueueAutoAck() {
				c.metric.ObserveAck(true)
				continue
			}

			switch acknowledgement.Acknowledgement {
			case Ack:
				err := d.Ack(false)
				if err != nil {
					c.metric.ObserveAck(false)
					c.logger.Error("failed to ack message", zap.Error(err))

					if c.handler.MustStopOnAckError() {
						return stacktrace.Propagate(err, "stop consuming due to ack error")
					}
					continue
				}

				c.metric.ObserveAck(true)
				c.logger.Info("successful ack message")
				continue
			case Nack:
				err := d.Nack(false, acknowledgement.Requeue)
				if err != nil {
					c.metric.ObserveNack(false)
					c.logger.Error("failed to nack message", zap.Error(err))

					if c.handler.MustStopOnNAckError() {
						return stacktrace.Propagate(err, "stop consuming due to nack error")
					}
					continue
				}
				c.metric.ObserveNack(true)
				c.logger.Info("successful nack message")
				continue
			case Reject:
				err := d.Reject(acknowledgement.Requeue)
				if err != nil {
					c.metric.ObserveReject(false)
					c.logger.Error("failed to reject message", zap.Error(err))

					if c.handler.MustStopOnRejectError() {
						return stacktrace.Propagate(err, "stop consuming due to reject error")
					}
					continue
				}
				c.metric.ObserveReject(true)
				c.logger.Info("successful rejected message")
				continue
			default:
				return stacktrace.NewError("acknowledgement type not in predefined")
			}
		case notifyErr := <-c.client.notify:
			return stacktrace.NewError("rabbitmq notified it stopped working, reason %s", notifyErr.Reason)
		}
	}
}
