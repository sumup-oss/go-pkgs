package rabbitmq

import (
	"github.com/palantir/stacktrace"

	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"

	"github.com/streadway/amqp"
)

func NewRabbitMQConsumer(
	client *RabbitMQClient,
	handler Handler,
	logger logger.StructuredLogger, // decide if we want it, maybe not just return errors
	metric Metric,
) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		client:  client,
		done:    make(chan error),
		handler: handler,
		logger:  logger,
		metric:  metric,
	}
}

type RabbitMQConsumer struct {
	client  *RabbitMQClient
	done    chan error
	handler Handler
	logger  logger.StructuredLogger
	metric  Metric
}

func (c *RabbitMQConsumer) Run(shutdownChan <-chan struct{}) error {

	go func() {
		select {
		case <-shutdownChan:
			c.logger.Info("Received shutdown signal. Going to close rabbit connections.")
			c.client.channel.Cancel(c.handler.GetConsumerTag(), false) //this will wait for the consumer to finish
			err := <-c.done
			c.logger.Error("cancel error", zap.Error(err))
			c.client.Shutdown()

			return
		}
	}()

	deliveries, err := c.client.channel.Consume(
		c.handler.GetQueue(),
		c.handler.GetConsumerTag(),
		false,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return stacktrace.Propagate(err, "couldn't start consuming from channel")
	}

	err = c.handle(deliveries, c.done)

	return err
}

func (c *RabbitMQConsumer) handle(deliveries <-chan amqp.Delivery, done chan error) error {
	for d := range deliveries {
		c.logger.Debug(
			"msg delivered",
			zap.Uint64("tag", d.DeliveryTag),
			zap.ByteString("body", d.Body),
		)

		ack, err := c.handler.ReceiveMessage(d.Body)
		if !ack {

		}

		err = d.Ack(false)
		if err != nil {
			c.logger.Error("failed to ack message", zap.Error(err))
		}
		c.logger.Error("successful ack message")
	}
	c.logger.Info("handle: deliveries channel closed")
	done <- stacktrace.NewError("foo bar")
	return stacktrace.NewError("foo bar")
}
