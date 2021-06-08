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
	"sync"

	"github.com/palantir/stacktrace"

	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"

	"github.com/streadway/amqp"
)

type ConsumerConfig struct {
	// PrefetchCount configures how many in-flight "deliveries" are available to the consumer to ack/nack.
	// ref: https://www.rabbitmq.com/consumer-prefetch.html
	// There's no default value for the reason that it's very easy to misuse RMQ and have multiple consumers
	// with too many deliveries in flight which results into badly distributed work load and high memory footprint
	// of the consumers.
	PrefetchCount int
}

type Consumer struct {
	client  RabbitMQClientInterface
	handler Handler
	logger  logger.StructuredLogger
	metric  Metric
	cfg     ConsumerConfig
	stopWg  sync.WaitGroup
}

func NewConsumer(
	client RabbitMQClientInterface,
	handler Handler,
	logger logger.StructuredLogger,
	metric Metric,
	cfg ConsumerConfig,
) *Consumer {
	return &Consumer{
		client:  client,
		handler: handler,
		logger:  logger,
		metric:  metric,
		cfg:     cfg,
		stopWg:  sync.WaitGroup{},
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	channel, err := c.client.CreateChannel(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create a RMQ channel")
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	closeCh := channel.NotifyClose(make(chan *amqp.Error))

	go func() {
		select {
		case rmqErr := <-closeCh:
			cancelFunc()
			c.logger.Warn(
				"RMQ closed the connection",
				zap.String("reason", rmqErr.Reason),
				zap.Int("code", rmqErr.Code),
				zap.Bool("recover", rmqErr.Recover),
				zap.Bool("server", rmqErr.Server),
			)
			return
		case <-ctx.Done():
			c.logger.Info("Received context cancel. Going to close RMQ connections.")
			err = channel.Cancel(c.handler.GetConsumerTag(), false)
			if err != nil {
				c.logger.Warn("failed to cancel the RMQ channel while stopping handler", logger.ErrorField(err))
			}

			// NOTE: We must process the events before we close the channel
			// otherwise we cant ACK/NACK.
			if c.handler.WaitToConsumeInflight() {
				c.stopWg.Wait()
			}

			_ = channel.Close()

			c.logger.Info("RMQ consumer stopped.")
			_ = c.client.Close()
		}
	}()

	if ctx.Err() != nil {
		return stacktrace.Propagate(ctx.Err(), "context canceled")
	}

	err = channel.Qos(c.cfg.PrefetchCount, 0, false)
	if err != nil {
		return stacktrace.Propagate(err, "failed to set RMQ channel's QoS prefetch count to: %d", c.cfg.PrefetchCount)
	}

	deliveries, err := channel.Consume(
		c.handler.GetQueueName(),
		c.handler.GetConsumerTag(),
		c.handler.QueueAutoAck(),
		c.handler.ExclusiveConsumer(),
		false,
		false,
		nil,
	)
	if err != nil {
		return stacktrace.Propagate(err, "couldn't start consuming from RMQ channel")
	}

	err = c.handleDeliveries(ctx, deliveries)
	return stacktrace.Propagate(err, "failed/stopped handling RMQ consumer deliveries")
}

// nolint:gocognit
func (c *Consumer) handleDeliveries(
	ctx context.Context,
	deliveries <-chan amqp.Delivery,
) error {
	for {
		select {
		case <-ctx.Done():
			c.logger.Warn("RMQ handler stopping")
			return ctx.Err()
		case d, hasMore := <-deliveries:
			if !hasMore {
				c.logger.Warn("RMQ handler deliveries channel closed.")
				return stacktrace.NewError("RMQ handler deliveries channel closed.")
			}

			// TODO: Add option to parallelize processing
			c.stopWg.Add(1)
			err := c.handleSingleDelivery(ctx, &d)
			c.stopWg.Done()
			if err != nil {
				return stacktrace.Propagate(err, "failed to process RMQ delivery")
			}
		}
	}
}

func tracingField(d *amqp.Delivery) zap.Field {
	if d.CorrelationId == "" {
		return zap.Skip()
	}

	return zap.String("tracing_id", d.CorrelationId)
}

func (c *Consumer) handleSingleDelivery(ctx context.Context, d *amqp.Delivery) error {
	c.metric.ObserveMsgDelivered()

	acknowledgement, err := c.handler.ReceiveMessage(ctx, &Message{
		Body:          d.Body,
		CorrelationID: d.CorrelationId,
	})
	if err != nil {
		return stacktrace.Propagate(err, "handler returned error")
	}

	if c.handler.QueueAutoAck() {
		c.metric.ObserveAck(true)
		return nil
	}

	switch acknowledgement.Acknowledgement {
	case Ack:
		err := d.Ack(false)
		if err != nil {
			c.metric.ObserveAck(false)
			c.logger.Error(
				"failed to ack message",
				zap.Error(err),
				tracingField(d),
			)

			if c.handler.MustStopOnAckError() {
				return stacktrace.Propagate(err, "stop consuming due to ack error")
			}

			return nil
		}

		c.metric.ObserveAck(true)
		c.logger.Info(
			"successful ack message",
			tracingField(d),
		)
		return nil
	case Nack:
		err := d.Nack(false, acknowledgement.Requeue)
		if err != nil {
			c.metric.ObserveNack(false)
			c.logger.Error(
				"failed to nack message",
				zap.Error(err),
				tracingField(d),
			)

			if c.handler.MustStopOnNAckError() {
				return stacktrace.Propagate(err, "stop consuming due to nack error")
			}

			return nil
		}

		c.metric.ObserveNack(true)
		c.logger.Info(
			"successful nack message",
			tracingField(d),
		)

		return nil
	case Reject:
		err := d.Reject(acknowledgement.Requeue)
		if err != nil {
			c.metric.ObserveReject(false)
			c.logger.Error(
				"failed to reject message",
				zap.Error(err),
				tracingField(d),
			)

			if c.handler.MustStopOnRejectError() {
				return stacktrace.Propagate(err, "stop consuming due to reject error")
			}

			return nil
		}
		c.metric.ObserveReject(true)
		c.logger.Info(
			"successful rejected message",
			tracingField(d),
		)

		return nil
	default:
		return stacktrace.NewError("acknowledgement type not in predefined")
	}
}
