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

type Metric interface { //nolint:interfacebloat
	ObserveRabbitMQConnectionFailed()
	ObserveRabbitMQConnectionRetry()
	ObserveRabbitMQConnection()

	ObserveRabbitMQChanelConnectionFailed()
	ObserveRabbitMQChanelConnectionRetry()
	ObserveRabbitMQChanelConnection()

	ObserveMsgDelivered()
	ObserveAck(success bool)
	ObserveNack(success bool)
	ObserveReject(success bool)
	ObserveMsgPublish(success bool)
}

type NullMetric struct{}

func (n *NullMetric) ObserveRabbitMQConnectionFailed()       {}
func (n *NullMetric) ObserveRabbitMQConnectionRetry()        {}
func (n *NullMetric) ObserveRabbitMQConnection()             {}
func (n *NullMetric) ObserveRabbitMQChanelConnectionFailed() {}
func (n *NullMetric) ObserveRabbitMQChanelConnectionRetry()  {}
func (n *NullMetric) ObserveRabbitMQChanelConnection()       {}
func (n *NullMetric) ObserveMsgDelivered()                   {}
func (n *NullMetric) ObserveAck(success bool)                {}
func (n *NullMetric) ObserveNack(success bool)               {}
func (n *NullMetric) ObserveReject(success bool)             {}
func (n *NullMetric) ObserveMsgPublish(success bool)         {}
