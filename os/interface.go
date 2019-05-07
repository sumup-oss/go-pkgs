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
		Chdir(dir string) error
		Getwd() (string, error)
		Mkdir(dirname string, perm os.FileMode) error
		MkdirAll(dirname string, perm os.FileMode) error
		Exit(statusCode int)
		Stderr() io.Writer
		Stdin() io.Reader
		Stdout() io.Writer
		Args() []string
		Stat(filepath string) (os.FileInfo, error)
		IsNotExist(err error) bool
		OpenFile(path string, flag int, perm os.FileMode) (*os.File, error)
		ReadFile(filename string) ([]byte, error)
		ReadDir(dirname string) ([]os.FileInfo, error)
		WriteFile(path string, data []byte, perm os.FileMode) error
		ExpandTilde(path string) (string, error)
		Getenv(key string) string
		GetOS() string
		ExecuteWithStreams(cmd string, arg, env []string, dir string, stdout, stderr io.Writer) error
		ResolvePath(path string) (string, error)
		Remove(path string) error
		RemoveAll(path string) error
		CurrentUser() (*user.User, error)
		Create(name string) (*os.File, error)
		IsDir(path string) error
		IsFile(path string) error
		TempDir(dir, prefix string) (string, error)
		TempFile(dir, pattern string) (*os.File, error)
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
