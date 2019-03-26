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
	"bytes"
	"regexp"

	"github.com/sumup-oss/go-pkgs/logger"
)

var removeColorRegex = regexp.MustCompile(`\x1B\[\d+m`)

type RealtimeWriter struct {
	buffer    *bytes.Buffer
	logBuffer *bytes.Buffer
	logger    logger.Logger
	logLevel  logger.Level
}

func NewRealtimeWriter(log logger.Logger, logLevel logger.Level) *RealtimeWriter {
	buffer := make([]byte, 0, 20)
	logBuffer := make([]byte, 0)
	return &RealtimeWriter{
		buffer:    bytes.NewBuffer(buffer),
		logBuffer: bytes.NewBuffer(logBuffer),
		logger:    log,
		logLevel:  logLevel,
	}
}

func (writer *RealtimeWriter) Write(p []byte) (n int, err error) {
	writer.log(p)
	written, err := writer.buffer.Write(p)
	return written, err
}

func (writer *RealtimeWriter) log(p []byte) {
	for _, b := range p {
		if b == '\n' {
			writer.logger.Logf(writer.logLevel, writer.logBuffer.String())
			writer.logBuffer.Reset()
			continue
		}

		writer.logBuffer.WriteByte(b)
	}
}

func (writer *RealtimeWriter) GetOutput() string {
	output := writer.buffer.String()
	return removeColorRegex.ReplaceAllString(output, ``)
}
