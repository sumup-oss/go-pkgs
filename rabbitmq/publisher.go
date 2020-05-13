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
	"github.com/streadway/amqp"

	"github.com/sumup-oss/go-pkgs/logger"
)

type RabbitMQPublisher struct {
	client *RabbitMQClient
	logger logger.StructuredLogger
	metric Metric
}

func NewPublisher(client *RabbitMQClient, logger logger.StructuredLogger, metric Metric) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		client: client,
		logger: logger,
		metric: metric,
	}
}

func (p *RabbitMQPublisher) Publish(
	exchange,
	key string,
	mandatory,
	immediate bool,
	expiration string,
	body []byte,
	args amqp.Table,
) error {
	err := p.client.channel.Publish(
		exchange,
		key,
		mandatory,
		immediate,
		amqp.Publishing{
			Headers:    args,
			Expiration: expiration,
			Body:       body,
		},
	)

	p.metric.ObserveMsgPublish(err == nil)

	return err
}

func (p *RabbitMQPublisher) Close() error {
	return p.client.Close()
}
