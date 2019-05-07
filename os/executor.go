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

package os

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	//nolint:goimports
	"github.com/mattes/go-expand-tilde"
	"github.com/palantir/stacktrace"
)

// Compile-time proof of interfaces implementation.
var _ OsExecutor = (*RealOsExecutor)(nil)
var _ CommandExecutor = (*RealOsExecutor)(nil)
var _ EnvProvider = (*RealOsExecutor)(nil)
var _ IOStreamsProvider = (*RealOsExecutor)(nil)

type RealOsExecutor struct{}

func (ex *RealOsExecutor) Chdir(dir string) error {
	return osChdir(dir)
}

func (ex *RealOsExecutor) Getwd() (string, error) {
	return osGetwd()
}

func (ex *RealOsExecutor) Mkdir(dirname string, perm os.FileMode) error {
	return osMkdir(dirname, perm)
}

func (ex *RealOsExecutor) MkdirAll(path string, perm os.FileMode) error {
	return osMkdirAll(path, perm)
}

func (ex *RealOsExecutor) Exit(statusCode int) {
	osExit(statusCode)
}

func (ex *RealOsExecutor) Stderr() io.Writer {
	return osStderr
}

func (ex *RealOsExecutor) Stdin() io.Reader {
	return osStdin
}

func (ex *RealOsExecutor) Stdout() io.Writer {
	return osStdout
}

func (ex *RealOsExecutor) Args() []string {
	return osArgs
}

func (ex *RealOsExecutor) Stat(filename string) (os.FileInfo, error) {
	return osStat(filename)
}

func (ex *RealOsExecutor) IsNotExist(err error) bool {
	return osIsNotExist(err)
}

func (ex *RealOsExecutor) OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	return osOpenfile(path, flag, perm)
}

func (ex *RealOsExecutor) WriteFile(path string, data []byte, perm os.FileMode) error {
	return ioutilWriteFile(path, data, perm)
}

func (ex *RealOsExecutor) ExpandTilde(path string) (string, error) {
	return tilde.Expand(path)
}

func (ex *RealOsExecutor) Getenv(key string) string {
	return osGetenv(key)
}

func (ex *RealOsExecutor) GetOS() string {
	return runtime.GOOS
}

func (ex *RealOsExecutor) Execute(
	cmd string,
	arg []string,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	err := ex.ExecuteWithStreams(cmd, arg, env, dir, &stdout, &stderr)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (ex *RealOsExecutor) ExecuteWithStreams(
	cmd string,
	arg []string,
	env []string,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	command := execCommand(cmd, arg...)

	if len(env) > 0 {
		command.Env = env
	}

	command.Stdout = stdout
	command.Stderr = stderr
	command.Dir = dir

	err := command.Run()
	return stacktrace.Propagate(err, "executing command failed")
}

func (ex *RealOsExecutor) ResolvePath(path string) (string, error) {
	expandedPath, err := ex.ExpandTilde(path)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"cannot expand (resolve tilde `~` and similar)",
		)
	}

	path, err = filepath.Abs(expandedPath)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"cannot get absolute path of expanded path %s",
			expandedPath,
		)
	}

	return path, nil
}

func (ex *RealOsExecutor) Remove(path string) error {
	return osRemove(path)
}

func (ex *RealOsExecutor) CurrentUser() (*user.User, error) {
	return userCurrent()
}

func (ex *RealOsExecutor) Create(name string) (*os.File, error) {
	return osCreate(name)
}

func (ex *RealOsExecutor) ReadFile(filename string) ([]byte, error) {
	return ioutilReadFile(filename)
}

func (ex *RealOsExecutor) IsDir(path string) error {
	fileInfo, err := ex.Stat(path)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return errors.New("not a dir")
	}

	return nil
}

func (ex *RealOsExecutor) IsFile(path string) error {
	fileInfo, err := ex.Stat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return errors.New("not a file")
	}

	return nil
}

func (ex *RealOsExecutor) RemoveAll(path string) error {
	return osRemoveAll(path)
}

func (ex *RealOsExecutor) TempDir(dir, prefix string) (name string, err error) {
	return ioutilTempDir(dir, prefix)
}

func (ex *RealOsExecutor) TempFile(dir, pattern string) (f *os.File, err error) {
	return ioutilTempFile(dir, pattern)
}

func (ex *RealOsExecutor) ReadDir(dirname string) ([]os.FileInfo, error) {
	return ioutilReadDir(dirname)
}
