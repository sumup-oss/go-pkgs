package logger

import (
	gsyslog "github.com/hashicorp/go-syslog"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
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

func (core *ZapSyslogCore) With(fields []zapcore.Field) zapcore.Core {
	clone := core.clone()
	for _, field := range fields {
		field.AddTo(clone.encoder)
	}

	return clone
}

// NOTE: We pass `entry` by value to satisfy the interface requirements
// nolint:gocritic
func (core *ZapSyslogCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checked.AddCore(entry, core)
	}

	return checked
}

// NOTE: We pass `entry` by value to satisfy the interface requirements
// nolint:gocritic
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
	default:
		return errors.Errorf("unknown log level: %v", entry.Level)
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
