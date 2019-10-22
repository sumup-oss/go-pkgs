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

package executor

import (
	"context"
	"strings"

	"github.com/sumup-oss/go-pkgs/logger"
	"github.com/sumup-oss/go-pkgs/os"
)

var _ os.OsExecutor = (*ExecuteLogger)(nil)

// ExecuteLogger is os.OsExecutor decorator, that decorates the Execute method for real time debug logging.
type ExecuteLogger struct {
	os.OsExecutor

	log      logger.Logger
	logLevel logger.Level
}

func NewExecuteLogger(osExecutor os.OsExecutor, log logger.Logger) *ExecuteLogger {
	return NewExecuteLoggerWithLogLevel(osExecutor, log, logger.DebugLevel)
}

func NewExecuteLoggerWithLogLevel(osExecutor os.OsExecutor, log logger.Logger, logLevel logger.Level) *ExecuteLogger {
	return &ExecuteLogger{
		OsExecutor: osExecutor,
		log:        log,
		logLevel:   logLevel,
	}
}

func (c *ExecuteLogger) Execute(cmd string, arg []string, env []string, dir string) ([]byte, []byte, error) {
	c.log.Debugf("command# %s %s", cmd, strings.Join(arg, " "))

	stdout := NewRealtimeWriter(c.log, c.logLevel)
	stderr := NewRealtimeWriter(c.log, c.logLevel)

	err := c.ExecuteWithStreams(cmd, arg, env, dir, stdout, stderr)

	return []byte(stdout.GetOutput()), []byte(stderr.GetOutput()), err
}

func (c *ExecuteLogger) ExecuteContext(
	ctx context.Context,
	cmd string,
	arg []string,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	c.log.Debugf("command# %s %s", cmd, strings.Join(arg, " "))

	stdout := NewRealtimeWriter(c.log, c.logLevel)
	stderr := NewRealtimeWriter(c.log, c.logLevel)

	err := c.ExecuteWithStreamsContext(ctx, cmd, arg, env, dir, stdout, stderr)

	return []byte(stdout.GetOutput()), []byte(stderr.GetOutput()), err
}
