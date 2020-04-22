package rabbitmq

import "github.com/streadway/amqp"

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ExchangeConfig struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type QueueBindConfig struct {
	Name     string
	Key      string
	Exchange string
	NoWait   bool
	Args     amqp.Table
}

type Setup struct {
	Exchanges     []ExchangeConfig
	Queues        []QueueConfig
	QueueBindings []QueueBindConfig
}
