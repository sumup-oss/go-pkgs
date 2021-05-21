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
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

type RetryableProducer struct {
	config        RetryableProducerConfig
	logger        logger.StructuredLogger
	metric        Metric
	clientFactory func(ctx context.Context, config *ClientConfig) (RabbitMQClientInterface, error)
	cancel        context.CancelFunc

	// mu protects the producer property
	mu       sync.RWMutex
	producer *Producer
}

type RetryableProducerConfig struct {
	RabbitClientConfig *ClientConfig
}

func NewRetryableProducer(
	newClientFactory func(ctx context.Context, config *ClientConfig) (RabbitMQClientInterface, error),
	config RetryableProducerConfig,
	logger logger.StructuredLogger,
	metric Metric,
) *RetryableProducer {
	ctx, cancel := context.WithCancel(context.Background())

	retryableProducer := &RetryableProducer{
		logger:        logger,
		metric:        metric,
		clientFactory: newClientFactory,
		config:        config,
		cancel:        cancel,
	}

	go retryableProducer.initProducer(ctx)

	return retryableProducer
}

func (p *RetryableProducer) Publish(
	exchange,
	key string,
	mandatory,
	immediate bool,
	expiration string,
	body []byte,
	args amqp.Table,
) error {
	p.mu.RLock()
	producer := p.producer
	p.mu.RUnlock()

	if producer == nil {
		return stacktrace.NewError("RabbitMQ Producer client not connected")
	}

	err := producer.Publish(exchange, key, mandatory, immediate, expiration, body, args)
	if err != nil {
		return stacktrace.Propagate(err, "failed to publish RMQ message")
	}

	return nil
}

func (p *RetryableProducer) newProducer(ctx context.Context) (*Producer, error) {
	if ctx.Err() != nil {
		p.logger.Info("received context cancel")
		return nil, nil
	}

	client, err := p.clientFactory(ctx, p.config.RabbitClientConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "RabbitMQ Failed to init client")
	}

	producer, err := NewProducer(client, p.logger, p.metric)
	if err != nil {
		connCloseErr := client.Close()
		p.logger.Error("cannot close RabbitMQ client connection", zap.Error(connCloseErr))
		return nil, stacktrace.Propagate(err, "RabbitMQ Failed to create new producer")
	}

	return producer, nil
}

func (p *RetryableProducer) initProducer(ctx context.Context) {
	for {
		producer, err := p.newProducer(ctx)
		if err != nil {
			p.logger.Info("failed to create producer with backoff", zap.Error(err))
			return
		}

		p.mu.Lock()
		p.producer = producer
		p.mu.Unlock()

		p.logger.Info("producer connected to RabbitMQ")

		select {
		case <-ctx.Done():
			p.logger.Info("received shut down signal")
			err := producer.Close()
			if err != nil {
				p.logger.Error("error when closing the producer", zap.Error(err))
			}
			return
		case <-producer.closeCh:
			p.logger.Info("RabbitMQ Producer Client closed the connection, trying to reconnect")
		}
	}
}

func (p *RetryableProducer) Close() {
	p.cancel()
}
