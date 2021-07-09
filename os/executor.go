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
	"context"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	tilde "github.com/mattes/go-expand-tilde"
	"github.com/palantir/stacktrace"
)

var (
	// Compile-time proof of interfaces implementation.
	_ OsExecutor        = (*RealOsExecutor)(nil)
	_ CommandExecutor   = (*RealOsExecutor)(nil)
	_ EnvProvider       = (*RealOsExecutor)(nil)
	_ IOStreamsProvider = (*RealOsExecutor)(nil)
)

type RealOsExecutor struct {
	stdErr io.Writer
	stdin  io.Reader
	stdout io.Writer
}

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

func (ex *RealOsExecutor) SetStderr(v io.Writer) {
	ex.stdErr = v
}

func (ex *RealOsExecutor) Stderr() io.Writer {
	if ex.stdErr == nil {
		return osStderr
	}

	return ex.stdErr
}

func (ex *RealOsExecutor) SetStdin(v io.Reader) {
	ex.stdin = v
}

func (ex *RealOsExecutor) Stdin() io.Reader {
	if ex.stdin == nil {
		return osStdin
	}

	return ex.stdin
}

func (ex *RealOsExecutor) SetStdout(v io.Writer) {
	ex.stdout = v
}

func (ex *RealOsExecutor) Stdout() io.Writer {
	if ex.stdout == nil {
		return osStdout
	}

	return ex.stdout
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
	arg,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	err := ex.ExecuteWithStreams(cmd, arg, env, dir, &stdout, &stderr)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (ex *RealOsExecutor) ExecuteContext(
	ctx context.Context,
	cmd string,
	arg,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	var stdout, stderr bytes.Buffer
	err := ex.ExecuteWithStreamsContext(ctx, cmd, arg, env, dir, &stdout, &stderr)

	return stdout.Bytes(), stderr.Bytes(), err
}

func (ex *RealOsExecutor) ExecuteWithStreams(
	cmd string,
	arg,
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

func (ex *RealOsExecutor) ExecuteWithStreamsContext(
	ctx context.Context,
	cmd string,
	arg,
	env []string,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	command := execCommandContext(ctx, cmd, arg...)

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

func (ex *RealOsExecutor) RemoveContents(path string, limit int) error {
	fd, err := ex.Open(path)
	if err != nil {
		return err
	}

	names, err := fd.Readdirnames(limit)
	if err != nil {
		return err
	}

	for _, name := range names {
		dirPath := filepath.Join(path, name)
		err = ex.RemoveAll(dirPath)
		if err == nil {
			continue
		}

		return err
	}

	return nil
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

func (ex *RealOsExecutor) Open(name string) (*os.File, error) {
	return osOpen(name)
}

func (ex *RealOsExecutor) Chmod(name string, mode os.FileMode) error {
	return osChmod(name, mode)
}

func (ex *RealOsExecutor) Readlink(name string) (string, error) {
	return osReadlink(name)
}

func (ex *RealOsExecutor) Symlink(oldname, newname string) error {
	return osSymlink(oldname, newname)
}

func (ex *RealOsExecutor) Lstat(name string) (os.FileInfo, error) {
	return osLstat(name)
}

func (ex *RealOsExecutor) CopyFile(src, dst string) error {
	srcfd, err := ex.Open(src)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open file: %s", src)
	}
	defer srcfd.Close()

	fileInfo, err := ex.Lstat(src)
	if err != nil {
		return stacktrace.Propagate(err, "failed to stat file: %s", src)
	}

	// NOTE: Golang does not have a `IsSymlink` check
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return ex.CopyLink(src, dst)
	}

	if !fileInfo.Mode().IsRegular() {
		return stacktrace.Propagate(err, "not a file: %s", src)
	}

	dstfd, err := ex.Create(dst)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create file: %s", dst)
	}
	defer dstfd.Close()

	_, err = io.Copy(dstfd, srcfd)
	if err != nil {
		return stacktrace.Propagate(err, "failed to copy file contents from %s to %s", src, dst)
	}

	err = ex.Chmod(dst, fileInfo.Mode())

	return stacktrace.Propagate(err, "failed to change file permissions, path %s", dst)
}

func (ex *RealOsExecutor) CopyLink(src, dst string) error {
	srcPath, err := ex.Readlink(src)
	if err != nil {
		return stacktrace.Propagate(err, "failed to readlink from src")
	}

	return ex.Symlink(srcPath, dst)
}

func (ex *RealOsExecutor) CopyDir(src, dst string) error {
	srcStat, err := ex.Stat(src)
	if err != nil {
		return stacktrace.Propagate(err, "failed to stat file: %s", src)
	}
	if !srcStat.IsDir() {
		return stacktrace.Propagate(err, "not a directory: %s", src)
	}

	err = ex.MkdirAll(dst, srcStat.Mode())
	if err != nil {
		return stacktrace.Propagate(err, "failed to create dirs to path: %s", dst)
	}

	fds, err := ex.ReadDir(src)
	if err != nil {
		return stacktrace.Propagate(err, "failed to read dir contents at path: %s", src)
	}

	for _, fd := range fds {
		srcPath := filepath.Join(src, fd.Name())
		dstPath := filepath.Join(dst, fd.Name())

		if fd.IsDir() {
			err = ex.CopyDir(srcPath, dstPath)
			if err != nil {
				return stacktrace.Propagate(
					err,
					"failed to copy dir from %s to %s",
					srcPath,
					dstPath,
				)
			}
		} else {
			err = ex.CopyFile(srcPath, dstPath)
			if err != nil {
				return stacktrace.Propagate(
					err,
					"failed to copy file from %s to %s",
					srcPath,
					dstPath,
				)
			}
		}
	}

	return nil
}

func (ex *RealOsExecutor) IsExist(err error) bool {
	return osIsExist(err)
}

func (ex *RealOsExecutor) Rename(oldPath, newPath string) error {
	return osRename(oldPath, newPath)
}
