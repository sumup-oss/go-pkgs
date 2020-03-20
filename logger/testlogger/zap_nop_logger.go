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

package testlogger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sumup-oss/go-pkgs/logger"
)

// Ensure that ZapNopLogger implements the StructuredLogger interface.
var _ logger.StructuredLogger = (*ZapNopLogger)(nil)

// ZapNopLogger is no-op StructuredLogger.
type ZapNopLogger struct {
	*zap.Logger
	level zapcore.Level
}

// NewZapNopLogger returns a no-op StructuredLogger.
func NewZapNopLogger() *ZapNopLogger {
	return &ZapNopLogger{
		Logger: zap.NewNop(),
		level:  zap.DebugLevel,
	}
}

func (z *ZapNopLogger) GetLevel() zapcore.Level {
	return z.level
}
