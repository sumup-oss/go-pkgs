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
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
)

// NOTE: Define standard library `os`, `exec` and `ioutil`
// builtin functions so that they can be mocked during test of `RealOsExecutor`.
// It's expected that the actual functionality is not tested, but only that they're invoked correctly.
// Testing the golang standard library seems redundant at this point.

var execCommand = exec.Command
var execCommandContext = exec.CommandContext
var ioutilReadDir = ioutil.ReadDir
var ioutilReadFile = ioutil.ReadFile
var ioutilTempDir = ioutil.TempDir
var ioutilTempFile = ioutil.TempFile
var ioutilWriteFile = ioutil.WriteFile
var osArgs = os.Args
var osChdir = os.Chdir
var osChmod = os.Chmod
var osCreate = os.Create
var osExit = os.Exit
var osGetenv = os.Getenv
var osSetenv = os.Setenv
var osGetwd = os.Getwd
var osIsExist = os.IsExist
var osIsNotExist = os.IsNotExist
var osLstat = os.Lstat
var osMkdir = os.Mkdir
var osMkdirAll = os.MkdirAll
var osOpen = os.Open
var osOpenfile = os.OpenFile
var osReadlink = os.Readlink
var osRemove = os.Remove
var osRemoveAll = os.RemoveAll
var osRename = os.Rename
var osStat = os.Stat
var osStderr = os.Stderr
var osStdin = os.Stdin
var osStdout = os.Stdout
var osSymlink = os.Symlink
var userCurrent = user.Current
