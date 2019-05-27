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
	"io"
	"os"
	"os/user"
)

type (
	OsExecutor interface {
		Args() []string
		Chdir(dir string) error
		Chmod(name string, mode os.FileMode) error
		CopyDir(src, dst string) error
		CopyFile(src, dst string) error
		CopyLink(src, dst string) error
		Create(name string) (*os.File, error)
		CurrentUser() (*user.User, error)
		ExecuteWithStreams(cmd string, arg, env []string, dir string, stdout, stderr io.Writer) error
		Exit(statusCode int)
		ExpandTilde(path string) (string, error)
		Getenv(key string) string
		GetOS() string
		Getwd() (string, error)
		IsDir(path string) error
		IsExist(err error) bool
		IsFile(path string) error
		IsNotExist(err error) bool
		Lstat(name string) (os.FileInfo, error)
		Mkdir(dirname string, perm os.FileMode) error
		MkdirAll(dirname string, perm os.FileMode) error
		Open(name string) (*os.File, error)
		OpenFile(path string, flag int, perm os.FileMode) (*os.File, error)
		ReadDir(dirname string) ([]os.FileInfo, error)
		ReadFile(filename string) ([]byte, error)
		Readlink(name string) (string, error)
		Remove(path string) error
		RemoveAll(path string) error
		RemoveContents(path string, limit int) error
		ResolvePath(path string) (string, error)
		SetStderr(v io.Writer)
		SetStdin(v io.Reader)
		SetStdout(v io.Writer)
		Stat(filepath string) (os.FileInfo, error)
		Stderr() io.Writer
		Stdin() io.Reader
		Stdout() io.Writer
		Symlink(oldname, newname string) error
		TempDir(dir, prefix string) (string, error)
		TempFile(dir, pattern string) (*os.File, error)
		WriteFile(path string, data []byte, perm os.FileMode) error
		CommandExecutor
	}
	CommandExecutor interface {
		Execute(cmd string, arg, env []string, dir string) ([]byte, []byte, error)
	}
	EnvProvider interface {
		Getenv(key string) string
		GetOS() string
	}
	IOStreamsProvider interface {
		Stderr() io.Writer
		Stdin() io.Reader
		Stdout() io.Writer
	}
)
