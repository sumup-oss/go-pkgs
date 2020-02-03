package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ErrorField creates a zap.Field for the corresponding error.
//
// The sole purpose of this func is to provide extensibility, on the way errors logging is handled.
// Currently it does not do anything different than zap.Error().
//
// Future version should wrap the error with an object that supports Errors interface:
// type Errors interface {
//   Errors() []error
// }
func ErrorField(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}

	return zap.Field{
		Key:       "error",
		Type:      zapcore.ErrorType,
		Interface: err,
	}
}
