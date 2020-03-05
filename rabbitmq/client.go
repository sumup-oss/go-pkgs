package rabbitmq

import (
	"time"

	"github.com/palantir/stacktrace"

	"github.com/streadway/amqp"
)

type ClientConfig struct {
	ConnectionURI         string
	Metric                Metric
	ConnectRetryAttempts  int
	InitialReconnectDelay time.Duration
}

type RabbitMQClient struct {
	amqpURI               string
	conn                  *amqp.Connection
	channel               *amqp.Channel
	metric                Metric
	connectRetryAttempts  int
	initialReconnectDelay time.Duration
}

func NewRabbitMQClient(config *ClientConfig) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		amqpURI:               config.ConnectionURI,
		metric:                config.Metric,
		connectRetryAttempts:  config.ConnectRetryAttempts,
		initialReconnectDelay: config.InitialReconnectDelay,
	}

	// dial rabbitmq
	var connError error
	for i := 0; i < config.ConnectRetryAttempts; i++ {
		conn, err := amqp.Dial(client.amqpURI)
		if err != nil {
			config.Metric.ObserveRabbitMQConnectionRetry()
			time.Sleep(client.initialReconnectDelay)
			connError = err
			continue
		}

		client.conn = conn
		client.metric.ObserveRabbitMQConnection()
		connError = nil
		break
	}

	if connError != nil {
		client.metric.ObserveRabbitMQConnectionFailed()
		return nil, stacktrace.Propagate(connError, "couldn't dial rabbitmq")
	}

	// create rabbitmq channel
	var channelError error
	for i := 0; i < client.connectRetryAttempts; i++ {
		channel, err := client.conn.Channel()
		if err != nil {
			client.metric.ObserveRabbitMQChanelConnectionRetry()
			time.Sleep(client.initialReconnectDelay)
			channelError = err
			continue
		}

		client.channel = channel
		client.metric.ObserveRabbitMQChanelConnection()
		channelError = nil

		break
	}

	if channelError != nil {
		client.metric.ObserveRabbitMQChanelConnectionFailed()
		return nil, stacktrace.Propagate(channelError, "couldn't create channel for rabbitmq")
	}

	return client, nil
}

func (c *RabbitMQClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return stacktrace.Propagate(err, "AMQP connection close")
	}

	return nil
}
