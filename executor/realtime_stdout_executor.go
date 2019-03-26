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
	"io"

	"github.com/sumup/go-pkgs/os"
)

var _ os.OsExecutor = (*RealtimeStdoutExecutor)(nil)

// RealtimeStdoutExecutor is os.OsExecutor decorator, that decorates the Execute method by writing
// executed commands stdout and stderr to executor's Stdout and Stderr.
type RealtimeStdoutExecutor struct {
	os.OsExecutor
}

// NewRealtimeStdoutExecutor creates RealtimeStdoutExecutor instance.
func NewRealtimeStdoutExecutor(osExecutor os.OsExecutor) *RealtimeStdoutExecutor {
	return &RealtimeStdoutExecutor{
		OsExecutor: osExecutor,
	}
}

// Execute executes a command.
func (executor *RealtimeStdoutExecutor) Execute(
	cmd string,
	arg []string,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	stdout := NewBufferedWriter(executor.Stdout())
	stderr := NewBufferedWriter(executor.Stderr())

	err := executor.ExecuteWithStreams(cmd, arg, env, dir, stdout, stderr)

	return stdout.Bytes(), stderr.Bytes(), err
}

// BufferedWriter is a writer that decorates an writer, by buffering a copy of all written bytes.
type BufferedWriter struct {
	writer io.Writer
	buffer *bytes.Buffer
}

// NewBufferedWriter creates BufferedWriter instance.
func NewBufferedWriter(writer io.Writer) *BufferedWriter {
	buffer := make([]byte, 0, 1024)

	return &BufferedWriter{
		writer: writer,
		buffer: bytes.NewBuffer(buffer),
	}
}

// Write writes data to the stream and appends the contents of data to the internal buffer.
func (w *BufferedWriter) Write(data []byte) (int, error) {
	written, err := w.buffer.Write(data)
	if err != nil {
		return written, err
	}

	_, err = w.writer.Write(data)
	return written, err
}

// Bytes returns the buffered data.
func (w *BufferedWriter) Bytes() []byte {
	return w.buffer.Bytes()
}
