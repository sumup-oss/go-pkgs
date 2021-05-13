package rabbitmq

import (
	"context"
	"sync"
	"time"

	"github.com/palantir/stacktrace"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

type PersistentProducer struct {
	producer          *Producer
	isClosed          uint32
	mu                sync.RWMutex
	reconnectTimeout  time.Duration
	rabbitClientCfg   *ClientConfig
	rabbitClientSetup *Setup
}

func NewPersistentProducer(
	client *RabbitMQClient,
	setup *Setup,
	logger logger.StructuredLogger,
	metric Metric,
	ctx context.Context,
	reconnectTimeout time.Duration,
) (*PersistentProducer, error) {
	producer, err := NewProducer(client, logger, metric)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create a persistent producer")
	}

	persistentProducer := &PersistentProducer{
		producer:          producer,
		reconnectTimeout:  reconnectTimeout,
		rabbitClientCfg:   client.cfg,
		rabbitClientSetup: setup,
	}

	go persistentProducer.unsafeReconnect(ctx)

	return persistentProducer, nil
}

func (p *PersistentProducer) Publish(
	exchange,
	key string,
	mandatory,
	immediate bool,
	expiration string,
	body []byte,
	args amqp.Table,
) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	err := p.producer.Publish(exchange, key, mandatory, immediate, expiration, body, args)
	if err != nil {
		if p.producer.isClosed {
			return stacktrace.Propagate(err, "connection to RabbitMQ client closed")
		}
		return stacktrace.Propagate(err, "failed to publish RMQ message")
	}

	return nil
}

func (p *PersistentProducer) unsafeReconnect(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// TODO: log that it's canceled
			p.producer.logger.Info("received shut down signal")
			return
		case <-p.producer.closeCh:
			p.producer.logger.Info("try to reconnect RabbitMQ producer")
			client, err := NewRabbitMQClient(ctx, p.rabbitClientCfg)
			if err != nil {
				p.producer.logger.Warn("RabbitMQ Failed to init client", zap.Error(err))
				time.Sleep(p.reconnectTimeout)
				continue
			}

			p.producer.logger.Info("created a new rabbit client")
			err = client.Setup(ctx, p.rabbitClientSetup)
			if err != nil {
				p.producer.logger.Warn("RabbitMQ Failed to setup client", zap.Error(err))
				time.Sleep(p.reconnectTimeout)
				continue
			}

			producer, err := NewProducer(client, p.producer.logger, p.producer.metric)
			if err != nil {
				p.producer.logger.Warn("RabbitMQ Failed to create new producer", zap.Error(err))
				time.Sleep(p.reconnectTimeout)
				continue
			}

			p.mu.Lock()
			p.producer = producer
			p.mu.Unlock()

			p.producer.logger.Info("successfully reconnected RabbitMQ producer")
		}
	}
}
