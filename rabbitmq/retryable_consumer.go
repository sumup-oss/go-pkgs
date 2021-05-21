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
	"time"

	"github.com/palantir/stacktrace"
	"go.uber.org/zap"

	"github.com/sumup-oss/go-pkgs/backoff"
	"github.com/sumup-oss/go-pkgs/logger"
)

type RetryableConsumer struct {
	config        RetryableConsumerConfig
	logger        logger.StructuredLogger
	metric        Metric
	handler       Handler
	clientFactory func(ctx context.Context, config *ClientConfig) (RabbitMQClientInterface, error)
}

type RetryableConsumerConfig struct {
	MaxRetryAttempts int
	// healthCheckFactor is a number representing how much N multiplied by backoffConfig.Max time is needed
	// for a block of code to run w/o returning an error, to consider it healthy.
	// E.g backConfig.Max = 1min, healthCheckFactor = 2, means that code needs to run 2min at least to be healthy
	// and retried again starting from backoffConfig.Base the next time it has an error.
	HealthCheckFactor  int
	BackoffConfig      *backoff.Config
	ConsumerConfig     ConsumerConfig
	RabbitClientConfig *ClientConfig
}

func NewRetryableConsumer(
	newClientFactory func(ctx context.Context, config *ClientConfig) (RabbitMQClientInterface, error),
	config RetryableConsumerConfig,
	logger logger.StructuredLogger,
	metric Metric,
	handler Handler,
) *RetryableConsumer {
	return &RetryableConsumer{
		clientFactory: newClientFactory,
		config:        config,
		handler:       handler,
		logger:        logger,
		metric:        metric,
	}
}

func (c *RetryableConsumer) Run(ctx context.Context) error {
	consumerBackoff := backoff.NewBackoff(c.config.BackoffConfig)
	currentRetryAttempts := 0

	for {
		startTime := time.Now()
		err := c.doRun(ctx)
		if err != nil {
			c.logger.Error("consumer run failed with error", zap.Error(err))

			if c.config.MaxRetryAttempts != 0 && currentRetryAttempts > c.config.MaxRetryAttempts {
				return stacktrace.NewError("retry attempts exceeded")
			}

			if time.Since(startTime) > time.Duration(c.config.HealthCheckFactor)*c.config.BackoffConfig.Max {
				consumerBackoff = backoff.NewBackoff(c.config.BackoffConfig)
				currentRetryAttempts = 0
			} else {
				currentRetryAttempts += 1
			}

			backoffDuration := consumerBackoff.Next()

			select {
			case <-ctx.Done():
				c.logger.Info("received context cancel")
				return nil
			case <-time.After(backoffDuration):
				continue
			}
		}

		return nil
	}
}

func (c *RetryableConsumer) doRun(ctx context.Context) error {
	c.logger.Info("RabbitMQ consumer Run")

	if ctx.Err() != nil {
		c.logger.Info("received context cancel")
		return nil
	}

	client, err := c.clientFactory(ctx, c.config.RabbitClientConfig)
	if err != nil {
		return stacktrace.Propagate(err, "RabbitMQ Failed to init client")
	}
	defer client.Close()

	c.logger.Info("starting to run the consumer")

	consumer := NewConsumer(client, c.handler, c.logger, c.metric, c.config.ConsumerConfig)
	err = consumer.Run(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "RabbitMQ consumer Run error")
	}

	return nil
}
