package rabbitmq

type Handler interface {
	GetQueue() string
	GetConsumerTag() string
	QueueAutoAck() bool
	ExclusiveConsumer() bool
	MustStopOnAckError() bool
	MustStopOnNAckError() bool
	MustStopOnRejectError() bool
	ReceiveMessage(payload []byte) (ack, nack, reject, requeue bool)
}
