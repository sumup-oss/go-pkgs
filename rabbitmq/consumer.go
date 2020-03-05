package rabbitmq

import (
	"github.com/palantir/stacktrace"

	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"

	"github.com/streadway/amqp"
)

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
	logger logger.StructuredLogger, // decide if we want it, maybe not just return errors
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

func (c *RabbitMQConsumer) Run(shutdownChan <-chan struct{}) error {
	go func() {
		<-shutdownChan
		c.logger.Info("Received shutdown signal. Going to close rabbit connections.")
		_ = c.client.channel.Cancel(c.handler.GetConsumerTag(), false) //this will wait for the consumer to finish

		<-c.done
		c.logger.Info("handler stopped")
		_ = c.client.Close()
	}()

	deliveries, err := c.client.channel.Consume(
		c.handler.GetQueue(),
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

	err = c.handle(deliveries, c.done)

	return stacktrace.Propagate(err, "stopped consumer")
}

func (c *RabbitMQConsumer) handle(deliveries <-chan amqp.Delivery, done chan bool) error {
	for d := range deliveries {
		c.logger.Debug(
			"msg delivered",
			zap.Uint64("tag", d.DeliveryTag),
			zap.ByteString("body", d.Body),
		)

		ack, nack, reject, requeue := c.handler.ReceiveMessage(d.Body)

		if c.handler.QueueAutoAck() {
			continue
		}

		if ack {
			err := d.Ack(false)
			if err != nil {
				c.logger.Error("failed to ack message", zap.Error(err))

				if c.handler.MustStopOnAckError() {
					return stacktrace.Propagate(err, "stop consuming due to ack error")
				}
			}
			c.logger.Error("successful ack message")
			continue
		}

		if nack {
			err := d.Nack(false, requeue)
			if err != nil {
				c.logger.Error("failed to nack message", zap.Error(err))

				if c.handler.MustStopOnNAckError() {
					return stacktrace.Propagate(err, "stop consuming due to nack error")
				}
			}
			c.logger.Error("successful nack message")
			continue
		}

		if reject {
			err := d.Reject(requeue)
			if err != nil {
				c.logger.Error("failed to reject message", zap.Error(err))

				if c.handler.MustStopOnRejectError() {
					return stacktrace.Propagate(err, "stop consuming due to reject error")
				}
			}
			c.logger.Error("successful rejected message")
			continue
		}
	}

	done <- true
	return nil
}
