package rabbitmq

import "go.uber.org/zap"

func tracingField(correlationID string) zap.Field {
	if correlationID == "" {
		return zap.Skip()
	}

	return zap.String("tracing_id", correlationID)
}
