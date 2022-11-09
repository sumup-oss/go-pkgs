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

package ostest

import (
	"context"
	"io"
	stdOs "os"
	"os/user"
	"testing"

	"github.com/sumup-oss/go-pkgs/os"

	"github.com/stretchr/testify/mock"
)

var _ os.OsExecutor = (*FakeOsExecutor)(nil)

type FakeOsExecutor struct {
	mock.Mock
}

// NewFakeOsExecutor creates FakeOsExecutor instance.
//
// nolint: thelper
func NewFakeOsExecutor(t *testing.T) *FakeOsExecutor {
	fake := &FakeOsExecutor{}
	fake.Test(t)

	return fake
}

func (f *FakeOsExecutor) Getwd() (string, error) {
	args := f.Called()

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) UserHomeDir() (string, error) {
	args := f.Called()

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) Chdir(dir string) error {
	args := f.Called(dir)

	return args.Error(0)
}

func (f *FakeOsExecutor) Mkdir(dirname string, perm stdOs.FileMode) error {
	args := f.Called(dirname, perm)

	return args.Error(0)
}

func (f *FakeOsExecutor) Execute(
	cmd string,
	arg []string,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	args := f.Called(cmd, arg, env, dir)
	rawStdout := args.Get(0)
	rawStderr := args.Get(1)
	returnErr := args.Error(2)

	var returnStdout, returnStderr []byte
	if rawStdout != nil {
		returnStdout = rawStdout.([]byte) // nolint:forcetypeassert
	}

	if rawStderr != nil {
		returnStderr = rawStderr.([]byte) // nolint:forcetypeassert
	}

	return returnStdout, returnStderr, returnErr
}

func (f *FakeOsExecutor) ExecuteContext(
	ctx context.Context,
	cmd string,
	arg []string,
	env []string,
	dir string,
) ([]byte, []byte, error) {
	args := f.Called(ctx, cmd, arg, env, dir)
	rawStdout := args.Get(0)
	rawStderr := args.Get(1)
	returnErr := args.Error(2)

	var returnStdout, returnStderr []byte
	if rawStdout != nil {
		returnStdout = rawStdout.([]byte) // nolint:forcetypeassert
	}

	if rawStderr != nil {
		returnStderr = rawStderr.([]byte) // nolint:forcetypeassert
	}

	return returnStdout, returnStderr, returnErr
}

func (f *FakeOsExecutor) MkdirAll(dirname string, perm stdOs.FileMode) error {
	args := f.Called(dirname, perm)

	return args.Error(0)
}

func (f *FakeOsExecutor) Exit(statusCode int) {
	f.Called(statusCode)
}

func (f *FakeOsExecutor) Stderr() io.Writer {
	args := f.Called()
	returnValue := args.Get(0)
	if returnValue == nil {
		return nil
	}

	return returnValue.(io.Writer)
}

func (f *FakeOsExecutor) Stdin() io.Reader {
	args := f.Called()
	returnValue := args.Get(0)
	if returnValue == nil {
		return nil
	}

	return returnValue.(io.Reader)
}

func (f *FakeOsExecutor) Stdout() io.Writer {
	args := f.Called()
	returnValue := args.Get(0)
	if returnValue == nil {
		return nil
	}

	return returnValue.(io.Writer)
}

func (f *FakeOsExecutor) Args() []string {
	args := f.Called()
	returnValue := args.Get(0)
	if returnValue == nil {
		return nil
	}

	return returnValue.([]string)
}

func (f *FakeOsExecutor) Stat(filepath string) (stdOs.FileInfo, error) {
	args := f.Called(filepath)
	returnValue := args.Get(0)
	err := args.Error(1)

	if returnValue == nil {
		return nil, err
	}

	return returnValue.(stdOs.FileInfo), err
}

func (f *FakeOsExecutor) IsNotExist(err error) bool {
	args := f.Called(err)

	return args.Bool(0)
}

func (f *FakeOsExecutor) OpenFile(path string, flag int, perm stdOs.FileMode) (*stdOs.File, error) {
	args := f.Called(path, flag, perm)

	returnValue := args.Get(0)
	err := args.Error(1)

	if returnValue == nil {
		return nil, err
	}

	return returnValue.(*stdOs.File), err
}

func (f *FakeOsExecutor) WriteFile(path string, data []byte, perm stdOs.FileMode) error {
	args := f.Called(path, data, perm)

	return args.Error(0)
}

func (f *FakeOsExecutor) ExpandTilde(path string) (string, error) {
	args := f.Called(path)

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) Getenv(key string) string {
	args := f.Called(key)

	return args.String(0)
}

func (f *FakeOsExecutor) Setenv(key, value string) error {
	args := f.Called(key, value)

	return args.Error(0)
}

func (f *FakeOsExecutor) GetOS() string {
	args := f.Called()

	return args.String(0)
}

func (f *FakeOsExecutor) ExecuteWithStreams(
	cmd string,
	arg []string,
	env []string,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	args := f.Called(cmd, arg, env, dir, stdout, stderr)

	return args.Error(0)
}

func (f *FakeOsExecutor) ExecuteWithStreamsContext(
	ctx context.Context,
	cmd string,
	arg []string,
	env []string,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	args := f.Called(ctx, cmd, arg, env, dir, stdout, stderr)

	return args.Error(0)
}

func (f *FakeOsExecutor) ResolvePath(path string) (string, error) {
	args := f.Called(path)

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) Remove(path string) error {
	args := f.Called(path)

	return args.Error(0)
}

func (f *FakeOsExecutor) CurrentUser() (*user.User, error) {
	args := f.Called()
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.(*user.User), err
}

func (f *FakeOsExecutor) Create(name string) (*stdOs.File, error) {
	args := f.Called(name)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.(*stdOs.File), err
}

func (f *FakeOsExecutor) ReadFile(filename string) (bytes []byte, e error) {
	args := f.Called(filename)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.([]byte), err
}

func (f *FakeOsExecutor) IsDir(path string) error {
	args := f.Called(path)

	return args.Error(0)
}

func (f *FakeOsExecutor) IsFile(path string) error {
	args := f.Called(path)

	return args.Error(0)
}

func (f *FakeOsExecutor) RemoveAll(path string) error {
	args := f.Called(path)

	return args.Error(0)
}

func (f *FakeOsExecutor) TempDir(dir, prefix string) (string, error) {
	args := f.Called(dir, prefix)

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) TempFile(dir, pattern string) (*stdOs.File, error) {
	args := f.Called(dir, pattern)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.(*stdOs.File), err
}

func (f *FakeOsExecutor) ReadDir(dirname string) ([]stdOs.FileInfo, error) {
	args := f.Called(dirname)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.([]stdOs.FileInfo), err
}

func (f *FakeOsExecutor) Chmod(name string, mode stdOs.FileMode) error {
	args := f.Called(name, mode)

	return args.Error(0)
}

func (f *FakeOsExecutor) CopyDir(src, dst string) error {
	args := f.Called(src, dst)

	return args.Error(0)
}

func (f *FakeOsExecutor) CopyFile(src, dst string) error {
	args := f.Called(src, dst)

	return args.Error(0)
}

func (f *FakeOsExecutor) CopyLink(src, dst string) error {
	args := f.Called(src, dst)

	return args.Error(0)
}

func (f *FakeOsExecutor) Open(name string) (*stdOs.File, error) {
	args := f.Called(name)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.(*stdOs.File), err
}

func (f *FakeOsExecutor) Readlink(name string) (string, error) {
	args := f.Called(name)

	return args.String(0), args.Error(1)
}

func (f *FakeOsExecutor) Symlink(oldname, newname string) error {
	args := f.Called(oldname, newname)

	return args.Error(0)
}

func (f *FakeOsExecutor) Lstat(name string) (stdOs.FileInfo, error) {
	args := f.Called(name)
	returnValue := args.Get(0)
	err := args.Error(1)
	if returnValue == nil {
		return nil, err
	}

	return returnValue.(stdOs.FileInfo), err
}

func (f *FakeOsExecutor) SetStderr(v io.Writer) {
	f.Called(v)
}

func (f *FakeOsExecutor) SetStdin(v io.Reader) {
	f.Called(v)
}

func (f *FakeOsExecutor) SetStdout(v io.Writer) {
	f.Called(v)
}

func (f *FakeOsExecutor) RemoveContents(path string, limit int) error {
	args := f.Called(path, limit)

	return args.Error(0)
}

func (f *FakeOsExecutor) IsExist(err error) bool {
	args := f.Called(err)

	return args.Bool(0)
}

func (f *FakeOsExecutor) Rename(oldPath, newPath string) error {
	args := f.Called(oldPath, newPath)

	return args.Error(0)
}

func (f *FakeOsExecutor) AppendToFile(path string, data []byte, perm stdOs.FileMode) error {
	args := f.Called(path, data, perm)

	return args.Error(0)
}
