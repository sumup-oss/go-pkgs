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

package logger

import (
	"os"

	gsyslog "github.com/hashicorp/go-syslog"

	"github.com/palantir/stacktrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Logger encoding types.
	EncodingJSON  = "json"
	EncodingPlain = "plain"

	// LogLevelPanic level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	LogLevelPanic = "PANIC"
	// LogLevelFatal level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	LogLevelFatal = "FATAL"
	// LogLevelError level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	LogLevelError = "ERROR"
	// LogLevelWarn level. Non-critical entries that deserve eyes.
	LogLevelWarn = "WARN"
	// LogInfoLevel level. General operational entries about what's going on inside the
	// application.
	LogLevelInfo = "INFO"
	// LogLevelDebug level. Usually only enabled when debugging. Very verbose logging.
	LogLevelDebug = "DEBUG"
)

var (
	zapLogLevels = map[string]zapcore.Level{
		LogLevelPanic: zapcore.PanicLevel,
		LogLevelFatal: zapcore.FatalLevel,
		LogLevelError: zapcore.ErrorLevel,
		LogLevelWarn:  zapcore.WarnLevel,
		LogLevelInfo:  zapcore.InfoLevel,
		LogLevelDebug: zapcore.DebugLevel,
	}

	defaultZapEncoderConfig = zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

type Configuration struct {
	Level         string
	Encoding      string
	StdoutEnabled bool
	SyslogEnabled bool
	// SyslogFacility is one of `KERN,USER,MAIL,DAEMON,AUTH,SYSLOG,LPR,NEWS,UUCP,CRON,AUTHPRIV,FTP,LOCAL0,
	// LOCAL1,LOCAL2,LOCAL3,LOCAL4,LOCAL5,LOCAL6,LOCAL7`
	SyslogFacility string
	// SyslogTag is tag for all messages produced
	SyslogTag string
	// Fields is an optional slice of fields which will be logged on each log invokation
	Fields []zapcore.Field
}

func NewZapLogger(config Configuration) (*ZapLogger, error) { //nolint:gocritic
	encoder, err := newEncoder(config.Encoding, &defaultZapEncoderConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "creating logger encoder failed")
	}

	level, err := getZapLevel(config.Level)
	if err != nil {
		return nil, stacktrace.Propagate(err, "creating logger failed")
	}

	var cores []zapcore.Core

	if config.StdoutEnabled {
		writer := zapcore.Lock(os.Stdout)
		cores = append(cores, zapcore.NewCore(encoder, writer, level))
	}

	if config.SyslogEnabled {
		// The syslog.LOG_INFO is used here intentionally,
		// since it is just a default severity if the syslog writer is used by its own.
		// All zapsyslog calls will overwrite this appropriately.
		// For example logger.Debug() will use syslog.LOG_DEBUG severity.
		writer, err := gsyslog.NewLogger(gsyslog.LOG_INFO, config.SyslogFacility, config.SyslogTag)
		if err != nil {
			return nil, stacktrace.Propagate(err, "creating syslog logging backend failed")
		}

		cores = append(cores, NewZapSyslogCore(level, encoder, writer))
	}

	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
	)

	if config.Fields != nil && len(config.Fields) > 0 {
		for _, f := range config.Fields {
			logger = logger.With(f)
		}
	}

	return &ZapLogger{
		Logger: logger,
		level:  level,
	}, nil
}

func newEncoder(encoding string, config *zapcore.EncoderConfig) (zapcore.Encoder, error) {
	switch encoding {
	case EncodingJSON:
		return zapcore.NewJSONEncoder(*config), nil
	case EncodingPlain:
		return zapcore.NewConsoleEncoder(*config), nil
	default:
		return nil, stacktrace.NewError("invalid encoder type: %s", encoding)
	}
}

func getZapLevel(level string) (zapcore.Level, error) {
	zapLevel, ok := zapLogLevels[level]
	if !ok {
		return zapcore.InfoLevel, stacktrace.NewError("invalid log level %s", level)
	}

	return zapLevel, nil
}

// Ensure that ZapLogger implements the StructuredLogger interface.
var _ StructuredLogger = (*ZapLogger)(nil)

type ZapLogger struct {
	*zap.Logger
	level zapcore.Level
}

func (z *ZapLogger) GetLevel() zapcore.Level {
	return z.level
}
