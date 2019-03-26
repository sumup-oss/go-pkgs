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
var osExit = os.Exit
var osChdir = os.Chdir
var osGetwd = os.Getwd
var osMkdir = os.Mkdir
var osMkdirAll = os.MkdirAll
var osStderr = os.Stderr
var osStdin = os.Stdin
var osStdout = os.Stdout
var osArgs = os.Args
var osStat = os.Stat
var osIsNotExist = os.IsNotExist
var osOpenfile = os.OpenFile
var ioutilWriteFile = ioutil.WriteFile
var osGetenv = os.Getenv
var osRemove = os.Remove
var osRemoveAll = os.RemoveAll
var userCurrent = user.Current
var osCreate = os.Create
var ioutilReadFile = ioutil.ReadFile
