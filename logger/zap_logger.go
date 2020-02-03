package logger

import (
	"log/syslog"
	"os"

	"github.com/palantir/stacktrace"
	"github.com/tchap/zapext/zapsyslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Logger encoding types
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
	syslogFacilities = map[int]syslog.Priority{
		0: syslog.LOG_LOCAL0,
		1: syslog.LOG_LOCAL1,
		2: syslog.LOG_LOCAL2,
		3: syslog.LOG_LOCAL3,
		4: syslog.LOG_LOCAL4,
		5: syslog.LOG_LOCAL5,
		6: syslog.LOG_LOCAL6,
		7: syslog.LOG_LOCAL7,
	}

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
	Level    string
	Encoding string

	StdoutEnabled bool

	SyslogEnabled  bool
	SyslogFacility int
	SyslogTag      string
}

type StructuredLogger interface {
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)

	Sync() error
}

// Ensure that zap.Logger implements the StructuredLogger interface.
var _ StructuredLogger = (*zap.Logger)(nil)

func NewZapLogger(config Configuration) (*zap.Logger, error) {
	cores := []zapcore.Core{}

	encoder, err := newEncoder(config.Encoding, &defaultZapEncoderConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "creating logger encoder failed")
	}

	level, err := getZapLevel(config.Level)
	if err != nil {
		return nil, stacktrace.Propagate(err, "creating logger failed")
	}

	if config.StdoutEnabled {
		writer := zapcore.Lock(os.Stdout)

		core := zapcore.NewCore(encoder, writer, level)
		cores = append(cores, core)
	}

	if config.SyslogEnabled {
		facility, err := getSyslogFacility(config.SyslogFacility)
		if err != nil {
			return nil, stacktrace.Propagate(err, "creating syslog logging backend failed")
		}

		// The syslog.LOG_INFO is used here intentionally,
		// since it is just a default severity if the syslog writer is used by its own.
		// All zapsyslog calls will overwrite this appropriately.
		// For example logger.Debug() will use syslog.LOG_DEBUG severity.
		writer, err := syslog.New(facility|syslog.LOG_INFO, config.SyslogTag)
		if err != nil {
			return nil, stacktrace.Propagate(err, "creating syslog logging backend failed")
		}

		core := zapsyslog.NewCore(level, encoder, writer)
		cores = append(cores, core)
	}

	logger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCallerSkip(2),
		zap.AddCaller(),
	)

	return logger, nil
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

func getSyslogFacility(facility int) (syslog.Priority, error) {
	priority, ok := syslogFacilities[facility]
	if !ok {
		return 0, stacktrace.NewError("invalid syslog facility %d, must be one of 0-7", priority)
	}

	return priority, nil
}
