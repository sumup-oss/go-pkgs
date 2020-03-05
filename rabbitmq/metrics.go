package rabbitmq

type Metric interface {
	ObserveRabbitMQConnectionFailed()
	ObserveRabbitMQConnectionRetry()
	ObserveRabbitMQConnection()

	ObserveRabbitMQChanelConnectionFailed()
	ObserveRabbitMQChanelConnectionRetry()
	ObserveRabbitMQChanelConnection()

	ObserveMsgDelivered()
	ObserveSuccessfulAck()
	ObserveFailedAck()
}

type NullMetric struct{}

func (n *NullMetric) ObserveRabbitMQConnectionFailed()       {}
func (n *NullMetric) ObserveRabbitMQConnectionRetry()        {}
func (n *NullMetric) ObserveRabbitMQConnection()             {}
func (n *NullMetric) ObserveRabbitMQChanelConnectionFailed() {}
func (n *NullMetric) ObserveRabbitMQChanelConnectionRetry()  {}
func (n *NullMetric) ObserveRabbitMQChanelConnection()       {}
func (n *NullMetric) ObserveMsgDelivered()                   {}
func (n *NullMetric) ObserveSuccessfulAck()                  {}
func (n *NullMetric) ObserveFailedAck()                      {}
