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

	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/logger"
)

type PersistentConsumer struct {
	consumer         *Consumer
	mu               sync.Mutex
	reconnectTimeout time.Duration
	rabbitClientCfg  *ClientConfig
}

func NewPersistentConsumer(
	client *RabbitMQClient,
	handler Handler,
	logger logger.StructuredLogger,
	metric Metric,
	cfg ConsumerConfig,
	reconnectTimeout time.Duration,
) *PersistentConsumer {
	consumer := NewConsumer(client, handler, logger, metric, cfg)

	return &PersistentConsumer{
		consumer:         consumer,
		mu:               sync.Mutex{},
		reconnectTimeout: reconnectTimeout,
		rabbitClientCfg:  client.cfg,
	}
}

func (c *PersistentConsumer) Run(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

StartEstablishConnection:
	c.consumer.logger.Info("RabbitMQ consumer Run StartEstablishConnection")

	client, err := NewRabbitMQClient(ctx, c.rabbitClientCfg)
	if err != nil {
		c.consumer.logger.Warn("RabbitMQ Failed to init client")

		select {
		case <-ctx.Done():
			c.consumer.logger.Info("received shut down signal")
			return nil
		case <-time.After(c.reconnectTimeout):
			goto StartEstablishConnection
		}
	}

	c.consumer.client = client

	select {
	case <-ctx.Done():
		c.consumer.logger.Info("received shut down signal")
		return nil
	default:
		c.consumer.logger.Info("try to Run the consumer")
		err := c.consumer.Run(ctx)
		if err != nil {
			c.consumer.logger.Error("RabbitMQ consumer Run error", zap.Error(err))

			select {
			case <-ctx.Done():
				c.consumer.logger.Info("received shut down signal")
				return nil
			case <-time.After(c.reconnectTimeout):
				goto StartEstablishConnection
			}
		}
	}

	return nil
}
