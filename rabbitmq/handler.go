package rabbitmq

type Handler interface {
	GetQueue() string
	GetConsumerTag() string
	QueueAutoAck() bool
	ExclusiveConsumer() bool
	ReceiveMessage(payload []byte) (ack, nack, reject, requeue bool)
}
