package rabbitmq

type Handler interface {
	GetQueue() string
	GetConsumerTag() string
	ReceiveMessage(payload []byte) (ack bool, err error)
}
