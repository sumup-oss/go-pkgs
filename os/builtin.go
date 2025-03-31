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
	"os"
	"os/exec"
	"os/user"
)

// NOTE: Define standard library `os` and `exec`.
// builtin functions so that they can be mocked during test of `RealOsExecutor`.
// It's expected that the actual functionality is not tested, but only that they're invoked correctly.
// Testing the golang standard library seems redundant at this point.

var (
	execCommand        = exec.Command
	execCommandContext = exec.CommandContext
	osReadDir          = os.ReadDir
	osReadFile         = os.ReadFile
	osTempDir          = os.MkdirTemp
	osCreateTemp       = os.CreateTemp
	osWriteFile        = os.WriteFile
	osArgs             = os.Args
	osChdir            = os.Chdir
	osUserHomeDir      = os.UserHomeDir
	osChmod            = os.Chmod
	osCreate           = os.Create
	osExit             = os.Exit
	osGetenv           = os.Getenv
	osSetenv           = os.Setenv
	osGetwd            = os.Getwd
	osIsExist          = os.IsExist
	osIsNotExist       = os.IsNotExist
	osLstat            = os.Lstat
	osMkdir            = os.Mkdir
	osMkdirAll         = os.MkdirAll
	osOpen             = os.Open
	osOpenfile         = os.OpenFile
	osReadlink         = os.Readlink
	osRemove           = os.Remove
	osRemoveAll        = os.RemoveAll
	osRename           = os.Rename
	osStat             = os.Stat
	osStderr           = os.Stderr
	osStdin            = os.Stdin
	osStdout           = os.Stdout
	osSymlink          = os.Symlink
	userCurrent        = user.Current
)
