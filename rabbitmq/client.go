package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/palantir/stacktrace"

	"github.com/streadway/amqp"
)

// ReconnectDelay ...
const ReconnectDelay = 2
const RetryAttempts = 5

type RabbitMqClientConfig struct {
	Host                  string
	Port                  int
	Username              string
	Password              string
	Metrics               Metric
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

func NewRabbitMQClient(config *RabbitMqClientConfig) (*RabbitMQClient, error) {
	amqpURI := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)

	client := &RabbitMQClient{
		amqpURI:               amqpURI,
		metric:                config.Metrics,
		connectRetryAttempts:  config.ConnectRetryAttempts,
		initialReconnectDelay: config.InitialReconnectDelay,
	}

	err := client.establishConnection()
	if err != nil {
		return nil, err
	}

	err = client.establishChannel()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *RabbitMQClient) establishConnection() error {
	for i := 0; i < c.connectRetryAttempts; i++ {
		conn, err := amqp.Dial(c.amqpURI)
		if err != nil {
			c.metric.ObserveRabbitMQConnectionRetry()
			time.Sleep(c.initialReconnectDelay)
			continue
		}

		c.conn = conn
		c.metric.ObserveRabbitMQConnection()

		return nil
	}

	c.metric.ObserveRabbitMQConnectionFailed()
	return stacktrace.NewError("couldn't dial rabbitmq")
}

func (c *RabbitMQClient) establishChannel() error {
	for i := 0; i < c.connectRetryAttempts; i++ {
		channel, err := c.conn.Channel()
		if err != nil {
			c.metric.ObserveRabbitMQChanelConnectionRetry()
			time.Sleep(c.initialReconnectDelay)
			continue
		}

		c.channel = channel
		c.metric.ObserveRabbitMQChanelConnection()

		return nil
	}

	c.metric.ObserveRabbitMQChanelConnectionFailed()
	return stacktrace.NewError("couldn't open channel for rabbitmq")
}

func (c *RabbitMQClient) Shutdown() error {
	err := c.conn.Close()
	if err != nil {
		return stacktrace.NewError("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return nil
}
