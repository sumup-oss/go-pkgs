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
	"fmt"
	"os"
	"sync"

	"github.com/sumup-oss/go-pkgs/logger"
)

var _ logger.Logger = (*TestLogger)(nil)

type TestLogger struct {
	mu        sync.Mutex
	Level     logger.Level
	InfoLogs  []string
	DebugLogs []string
	ErrorLogs []string
	WarnLogs  []string
}

func NewTestLogger(level logger.Level) *TestLogger {
	return &TestLogger{
		Level:     level,
		InfoLogs:  make([]string, 0),
		DebugLogs: make([]string, 0),
		ErrorLogs: make([]string, 0),
		WarnLogs:  make([]string, 0),
	}
}

// Taken from logrus implementation.
func (tl *TestLogger) sprintlnn(args ...interface{}) string {
	msg := fmt.Sprintln(args...)

	return msg[:len(msg)-1]
}

func (tl *TestLogger) SetLevel(level logger.Level) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.Level = level
}

func (tl *TestLogger) GetLevel() logger.Level {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	return tl.Level
}

func (tl *TestLogger) Print(args ...interface{}) {
	tl.Info(args...)
}

func (tl *TestLogger) Println(args ...interface{}) {
	tl.Infoln(args...)
}

func (tl *TestLogger) Printf(format string, args ...interface{}) {
	tl.Infof(format, args...)
}

func (tl *TestLogger) Info(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.InfoLevel {
		return
	}
	tl.InfoLogs = append(tl.InfoLogs, fmt.Sprint(args...))
}

func (tl *TestLogger) Infoln(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.InfoLevel {
		return
	}
	tl.InfoLogs = append(tl.InfoLogs, tl.sprintlnn(args...))
}

func (tl *TestLogger) Infof(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.InfoLevel {
		return
	}
	tl.InfoLogs = append(tl.InfoLogs, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Debug(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.DebugLogs = append(tl.DebugLogs, fmt.Sprint(args...))
}

func (tl *TestLogger) Debugln(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.DebugLogs = append(tl.DebugLogs, tl.sprintlnn(args...))
}

func (tl *TestLogger) Debugf(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.DebugLogs = append(tl.DebugLogs, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Warn(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.WarnLogs = append(tl.WarnLogs, fmt.Sprint(args...))
}

func (tl *TestLogger) Warnln(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.WarnLogs = append(tl.WarnLogs, tl.sprintlnn(args...))
}

func (tl *TestLogger) Warnf(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.DebugLevel {
		return
	}
	tl.WarnLogs = append(tl.WarnLogs, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Warning(args ...interface{}) {
	tl.Warn(args...)
}

func (tl *TestLogger) Warningln(args ...interface{}) {
	tl.Warnln(args...)
}

func (tl *TestLogger) Warningf(format string, args ...interface{}) {
	tl.Warnf(format, args...)
}

func (tl *TestLogger) Error(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.ErrorLevel {
		return
	}
	tl.ErrorLogs = append(tl.ErrorLogs, fmt.Sprint(args...))
}

func (tl *TestLogger) Errorln(args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.ErrorLevel {
		return
	}
	tl.ErrorLogs = append(tl.ErrorLogs, tl.sprintlnn(args...))
}

func (tl *TestLogger) Errorf(format string, args ...interface{}) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.Level < logger.ErrorLevel {
		return
	}
	tl.ErrorLogs = append(tl.ErrorLogs, fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Fatal(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
	os.Exit(1)
}

func (tl *TestLogger) Fatalln(args ...interface{}) {
	fmt.Fprint(os.Stderr, tl.sprintlnn(args...))
	os.Exit(1)
}

func (tl *TestLogger) Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func (tl *TestLogger) Panic(args ...interface{}) {
	panic(fmt.Sprint(args...))
}

func (tl *TestLogger) Panicln(args ...interface{}) {
	panic(tl.sprintlnn(args...))
}

func (tl *TestLogger) Panicf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (tl *TestLogger) Logf(level logger.Level, format string, args ...interface{}) {
	switch level {
	case logger.DebugLevel:
		tl.Debugf(format, args...)
	case logger.InfoLevel:
		tl.Infof(format, args...)
	case logger.WarnLevel:
		tl.Warnf(format, args...)
	case logger.ErrorLevel:
		tl.Errorf(format, args...)
	case logger.FatalLevel:
		tl.Fatalf(format, args...)
	case logger.PanicLevel:
		tl.Panicf(format, args...)
	default:
		panic(fmt.Sprintf("unknown log level: %s", level))
	}
}

func (tl *TestLogger) Log(level logger.Level, args ...interface{}) {
	switch level {
	case logger.DebugLevel:
		tl.Debug(args...)
	case logger.InfoLevel:
		tl.Info(args...)
	case logger.WarnLevel:
		tl.Warn(args...)
	case logger.ErrorLevel:
		tl.Error(args...)
	case logger.FatalLevel:
		tl.Fatal(args...)
	case logger.PanicLevel:
		tl.Panic(args...)
	default:
		panic(fmt.Sprintf("unknown log level: %s", level))
	}
}

func (tl *TestLogger) Logln(level logger.Level, args ...interface{}) {
	switch level {
	case logger.DebugLevel:
		tl.Debugln(args...)
	case logger.InfoLevel:
		tl.Infoln(args...)
	case logger.WarnLevel:
		tl.Warnln(args...)
	case logger.ErrorLevel:
		tl.Errorln(args...)
	case logger.FatalLevel:
		tl.Fatalln(args...)
	case logger.PanicLevel:
		tl.Panicln(args...)
	default:
		panic(fmt.Sprintf("unknown log level: %s", level))
	}
}
