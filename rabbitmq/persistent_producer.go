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
	"time"

	"github.com/palantir/stacktrace"
	"github.com/streadway/amqp"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

type PersistentProducer struct {
	producer          *Producer
	mu                sync.RWMutex
	reconnectTimeout  time.Duration
	rabbitClientCfg   *ClientConfig
	rabbitClientSetup *Setup
}

func NewPersistentProducer(
	ctx context.Context,
	client *RabbitMQClient,
	setup *Setup,
	logger logger.StructuredLogger,
	metric Metric,
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
			p.producer.logger.Info("received shut down signal")
			return
		case <-p.producer.closeCh:
			p.producer.logger.Info("try to reconnect RabbitMQ producer")
			client, err := NewRabbitMQClient(ctx, p.rabbitClientCfg)
			if err != nil {
				p.producer.logger.Warn("RabbitMQ Failed to init client", zap.Error(err))

				select {
				case <-ctx.Done():
					p.producer.logger.Info("received shut down signal")
					return
				case <-time.After(p.reconnectTimeout):
					continue
				}
			}

			producer, err := NewProducer(client, p.producer.logger, p.producer.metric)
			if err != nil {
				p.producer.logger.Warn("RabbitMQ Failed to create new producer", zap.Error(err))

				select {
				case <-ctx.Done():
					p.producer.logger.Info("received shut down signal")
					return
				case <-time.After(p.reconnectTimeout):
					continue
				}
			}

			p.mu.Lock()
			p.producer = producer
			p.mu.Unlock()

			p.producer.logger.Info("successfully reconnected RabbitMQ producer")
		}
	}
}
