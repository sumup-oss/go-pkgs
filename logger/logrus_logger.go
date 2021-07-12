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
	"io"
	"sync"

	//nolint:depguard
	"github.com/sirupsen/logrus"
)

var _ Logger = (*LogrusLogger)(nil)

type Level uint32

func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	}

	return "unknown"
}

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
}

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

type LogrusHook struct {
	Hook             Hook
	TriggerLevels    []Level
	calculatedLevels []logrus.Level
}

func (hook *LogrusHook) Fire(entry *logrus.Entry) error {
	basicEntry := &BasicEntry{
		Buffer:  entry.Buffer,
		Message: entry.Message,
		Level:   Level(uint32(entry.Level)),
		Time:    entry.Time,
		Fields:  make(map[string]interface{}),
	}

	return hook.Hook.Fire(basicEntry)
}

func (hook *LogrusHook) Levels() []logrus.Level {
	if len(hook.calculatedLevels) > 0 {
		return hook.calculatedLevels
	}

	hook.calculatedLevels = make([]logrus.Level, len(hook.TriggerLevels))
	for idx, lvl := range hook.TriggerLevels {
		hook.calculatedLevels[idx] = logrus.Level(lvl)
	}

	return hook.calculatedLevels
}

type LogrusLogger struct {
	internalLogger *logrus.Logger
	mu             *sync.Mutex
}

func NewLogrusLogger() *LogrusLogger {
	return &LogrusLogger{
		internalLogger: logrus.New(),
		mu:             &sync.Mutex{},
	}
}

// SetOutput sets the standard logger output.
func (std *LogrusLogger) SetOutput(out io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.internalLogger.Out = out
}

// SetLevel sets the standard logger level.
func (std *LogrusLogger) SetLevel(level Level) {
	std.mu.Lock()
	defer std.mu.Unlock()
	logrusLevel := logrus.Level(level)
	std.internalLogger.SetLevel(logrusLevel)
}

// GetLevel returns the standard logger level.
func (std *LogrusLogger) GetLevel() Level {
	std.mu.Lock()
	defer std.mu.Unlock()

	return Level(std.internalLogger.Level)
}

// AddHook adds a hook to the standard logger hooks.
func (std *LogrusLogger) AddHook(hook Hook) {
	std.mu.Lock()
	defer std.mu.Unlock()

	lrHook := &LogrusHook{
		TriggerLevels: hook.Levels(),
		Hook:          hook,
	}

	std.internalLogger.Hooks.Add(lrHook)
}

// Debug logs a message at level Debug on the standard logger.
func (std *LogrusLogger) Debug(args ...interface{}) {
	std.internalLogger.Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func (std *LogrusLogger) Print(args ...interface{}) {
	std.internalLogger.Print(args...)
}

// Info logs a message at level Info on the standard logger.
func (std *LogrusLogger) Info(args ...interface{}) {
	std.internalLogger.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warn(args ...interface{}) {
	std.internalLogger.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warning(args ...interface{}) {
	std.internalLogger.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func (std *LogrusLogger) Error(args ...interface{}) {
	std.internalLogger.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func (std *LogrusLogger) Panic(args ...interface{}) {
	std.internalLogger.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (std *LogrusLogger) Fatal(args ...interface{}) {
	std.internalLogger.Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func (std *LogrusLogger) Debugf(format string, args ...interface{}) {
	std.internalLogger.Debugf(format, args...)
}

// Logf logs a message at specified level on the standard logger.
func (std *LogrusLogger) Logf(level Level, format string, args ...interface{}) {
	std.internalLogger.Logf(logrus.Level(level), format, args...)
}

// Log logs a message at specified level on the standard logger.
func (std *LogrusLogger) Log(level Level, args ...interface{}) {
	std.internalLogger.Log(logrus.Level(level), args...)
}

// Logln logs a message at specified level on the standard logger.
func (std *LogrusLogger) Logln(level Level, args ...interface{}) {
	std.internalLogger.Logln(logrus.Level(level), args...)
}

// Printf logs a message at level Info on the standard logger.
func (std *LogrusLogger) Printf(format string, args ...interface{}) {
	std.internalLogger.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func (std *LogrusLogger) Infof(format string, args ...interface{}) {
	std.internalLogger.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warnf(format string, args ...interface{}) {
	std.internalLogger.Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warningf(format string, args ...interface{}) {
	std.internalLogger.Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func (std *LogrusLogger) Errorf(format string, args ...interface{}) {
	std.internalLogger.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func (std *LogrusLogger) Panicf(format string, args ...interface{}) {
	std.internalLogger.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (std *LogrusLogger) Fatalf(format string, args ...interface{}) {
	std.internalLogger.Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func (std *LogrusLogger) Debugln(args ...interface{}) {
	std.internalLogger.Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func (std *LogrusLogger) Println(args ...interface{}) {
	std.internalLogger.Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func (std *LogrusLogger) Infoln(args ...interface{}) {
	std.internalLogger.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warnln(args ...interface{}) {
	std.internalLogger.Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func (std *LogrusLogger) Warningln(args ...interface{}) {
	std.internalLogger.Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func (std *LogrusLogger) Errorln(args ...interface{}) {
	std.internalLogger.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func (std *LogrusLogger) Panicln(args ...interface{}) {
	std.internalLogger.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func (std *LogrusLogger) Fatalln(args ...interface{}) {
	std.internalLogger.Fatalln(args...)
}
