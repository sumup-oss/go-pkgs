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
	gsyslog "github.com/hashicorp/go-syslog"
	"go.uber.org/zap/zapcore"

	"github.com/sumup-oss/go-pkgs/errors"
)

type ZapSyslogCore struct {
	zapcore.LevelEnabler
	encoder zapcore.Encoder
	writer  gsyslog.Syslogger
}

func NewZapSyslogCore(enab zapcore.LevelEnabler, encoder zapcore.Encoder, writer gsyslog.Syslogger) *ZapSyslogCore {
	return &ZapSyslogCore{
		LevelEnabler: enab,
		encoder:      encoder,
		writer:       writer,
	}
}

func (core *ZapSyslogCore) With(fields []zapcore.Field) zapcore.Core { //nolint:ireturn
	clone := core.clone()
	for _, field := range fields {
		field.AddTo(clone.encoder)
	}

	return clone
}

// NOTE: We pass `entry` by value to satisfy the interface requirements
//
//nolint:gocritic
func (core *ZapSyslogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}

	return checked
}

// NOTE: We pass `entry` by value to satisfy the interface requirements
//
//nolint:gocritic
func (core *ZapSyslogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	buffer, err := core.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return errors.Wrap(err, "failed to encode log entry")
	}

	message := buffer.Bytes()

	switch entry.Level {
	case zapcore.DebugLevel:
		return core.writer.WriteLevel(gsyslog.LOG_DEBUG, message)
	case zapcore.InfoLevel:
		return core.writer.WriteLevel(gsyslog.LOG_INFO, message)
	case zapcore.WarnLevel:
		return core.writer.WriteLevel(gsyslog.LOG_WARNING, message)
	case zapcore.ErrorLevel:
		return core.writer.WriteLevel(gsyslog.LOG_ERR, message)
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return core.writer.WriteLevel(gsyslog.LOG_CRIT, message)
	case zapcore.InvalidLevel:
		return errors.New("invalid log level: %v", entry.Level)
	default:
		return errors.New("unknown log level: %v", entry.Level)
	}
}

func (core *ZapSyslogCore) Sync() error {
	return nil
}

func (core *ZapSyslogCore) clone() *ZapSyslogCore {
	return &ZapSyslogCore{
		LevelEnabler: core.LevelEnabler,
		encoder:      core.encoder.Clone(),
		writer:       core.writer,
	}
}
