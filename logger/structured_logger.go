package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StructuredLogger interface {
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	GetLevel() zapcore.Level
	Sync() error
}

// Ensure that StructuredNopLogger implements the StructuredLogger interface.
var _ StructuredLogger = (*StructuredNopLogger)(nil)

// StructuredNopLogger is no-op StructuredLogger.
type StructuredNopLogger struct {
	*zap.Logger
	level zapcore.Level
}

// NewStructuredNopLogger returns a no-op StructuredLogger.
//
// Note that if the passed level is not recognised it will default to INFO.
func NewStructuredNopLogger(level string) *StructuredNopLogger {
	zapLevel, ok := zapLogLevels[level]
	if !ok {
		zapLevel = zap.InfoLevel
	}

	return &StructuredNopLogger{
		Logger: zap.NewNop(),
		level:  zapLevel,
	}
}

func (z *StructuredNopLogger) GetLevel() zapcore.Level {
	return z.level
}
