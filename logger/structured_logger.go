// Copyright 2021 SumUp Ltd.
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

	// With creates a child logger and adds structured context to it. Fields added
	// to the child don't affect the parent, and vice versa.
	With(fields ...zap.Field) StructuredLogger

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
// Note that if the passed level is not recognized it will default to INFO.
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

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (z *StructuredNopLogger) With(fields ...zap.Field) StructuredLogger { //nolint:ireturn
	return &StructuredNopLogger{
		Logger: z.Logger.With(fields...),
		level:  z.level,
	}
}
