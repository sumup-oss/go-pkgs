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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealOsExecutor_Chdir(t *testing.T) {
	t.Run("it uses builtin `osChdir`", func(t *testing.T) {
		called := false
		var calledDirArgument string
		var calledErr error

		osChdir = func(dir string) error {
			called = true
			calledDirArgument = dir
			return calledErr
		}
		defer func() {
			osChdir = os.Chdir
		}()

		dirArgument := "/tmp/example"
		osExecutor := &RealOsExecutor{}

		err := osExecutor.Chdir(dirArgument)
		require.NoError(t, err)

		assert.True(t, called)
		assert.Equal(t, calledDirArgument, dirArgument)
	})
}

func TestRealOsExecutor_Getwd(t *testing.T) {
	t.Run("it uses builtin `osGetwd`", func(t *testing.T) {
		called := false
		calledDirArgument := "/tmp/mydir"
		var calledErr error

		osGetwd = func() (string, error) {
			called = true
			return calledDirArgument, calledErr
		}
		defer func() {
			osGetwd = os.Getwd
		}()

		osExecutor := &RealOsExecutor{}

		dir, err := osExecutor.Getwd()
		require.NoError(t, err)

		assert.True(t, called)
		assert.Equal(t, calledDirArgument, dir)
	})
}

func TestRealOsExecutor_Mkdir(t *testing.T) {
	t.Run("it uses builtin `Mkdir`", func(t *testing.T) {
		called := false
		var calledDirnameArgument string
		var calledPerm os.FileMode

		osMkdir = func(dirname string, perm os.FileMode) error {
			called = true
			calledDirnameArgument = dirname
			calledPerm = perm
			return nil
		}
		defer func() {
			osMkdir = os.Mkdir
		}()

		osExecutor := &RealOsExecutor{}

		dirArgument := "/tmp/example"
		permArgument := os.FileMode(0444)

		err := osExecutor.Mkdir(dirArgument, permArgument)
		require.NoError(t, err)

		assert.True(t, called)
		assert.Equal(t, calledDirnameArgument, dirArgument)
		assert.Equal(t, calledPerm, permArgument)
	})
}

func TestRealOsExecutor_MkdirAll(t *testing.T) {
	t.Run("it uses builtin `MkdirAll`", func(t *testing.T) {
		called := false
		var calledDirnameArgument string
		var calledPerm os.FileMode

		osMkdirAll = func(dirname string, perm os.FileMode) error {
			called = true
			calledDirnameArgument = dirname
			calledPerm = perm
			return nil
		}
		defer func() {
			osMkdirAll = os.MkdirAll
		}()

		osExecutor := &RealOsExecutor{}

		dirArgument := "/tmp/example"
		permArgument := os.FileMode(0444)

		err := osExecutor.MkdirAll(dirArgument, permArgument)
		require.NoError(t, err)

		assert.True(t, called)
		assert.Equal(t, calledDirnameArgument, dirArgument)
		assert.Equal(t, calledPerm, permArgument)
	})
}

func TestRealOsExecutor_Exit(t *testing.T) {
	t.Run("it uses builtin `osExit`", func(t *testing.T) {
		called := false
		var calledStatusCode int

		osExit = func(statusCode int) {
			called = true
			calledStatusCode = statusCode
		}
		defer func() {
			osExit = os.Exit
		}()

		osExecutor := &RealOsExecutor{}

		statusCodeArgument := 1

		osExecutor.Exit(statusCodeArgument)

		assert.True(t, called)
		assert.Equal(t, calledStatusCode, statusCodeArgument)
	})
}

func TestRealOsExecutor_Stderr(t *testing.T) {
	t.Run("it uses builtin `osStderr`", func(t *testing.T) {
		osStderr = &os.File{}

		osExecutor := &RealOsExecutor{}

		assert.Equal(t, osStderr, osExecutor.Stderr())
	})
}

func TestRealOsExecutor_Stdin(t *testing.T) {
	t.Run("it uses builtin `osStdin`", func(t *testing.T) {
		osStdin = &os.File{}

		osExecutor := &RealOsExecutor{}

		assert.Equal(t, osStdin, osExecutor.Stdin())
	})
}

func TestRealOsExecutor_Stdout(t *testing.T) {
	t.Run("it uses builtin `osStdout`", func(t *testing.T) {
		osStdout = &os.File{}

		osExecutor := &RealOsExecutor{}

		assert.Equal(t, osStdout, osExecutor.Stdout())
	})
}

func TestRealOsExecutor_Args(t *testing.T) {
	t.Run("it uses builtin `osArgs`", func(t *testing.T) {
		osArgs = []string{"1", "2", "3"}

		osExecutor := &RealOsExecutor{}

		actualOsArgs := osExecutor.Args()

		assert.Equal(t, osArgs, actualOsArgs)
	})
}

func TestRealOsExecutor_Stat(t *testing.T) {
	t.Run("it uses builtin `osStat`", func(t *testing.T) {
		called := false
		var calledFilename string

		osStat = func(filename string) (os.FileInfo, error) {
			called = true
			calledFilename = filename
			return nil, nil
		}
		defer func() {
			osStat = os.Stat
		}()

		osExecutor := &RealOsExecutor{}

		filenameArgument := "/tmp/myfile"
		actualFileInfo, actualErr := osExecutor.Stat(filenameArgument)
		require.NoError(t, actualErr)

		assert.True(t, called)
		assert.Equal(t, calledFilename, filenameArgument)
		assert.Nil(t, actualFileInfo)
	})
}

func TestRealOsExecutor_IsNotExist(t *testing.T) {
	t.Run("it uses builtin `osIsNotExist`", func(t *testing.T) {
		called := false
		var calledErr error
		calledReturn := true

		osIsNotExist = func(err error) bool {
			called = true
			calledErr = err
			return calledReturn
		}
		defer func() {
			osIsNotExist = os.IsNotExist
		}()

		osExecutor := &RealOsExecutor{}

		errArgument := errors.New("errargument")
		actualIsNotExist := osExecutor.IsNotExist(errArgument)

		assert.True(t, actualIsNotExist)
		assert.True(t, called)
		assert.Equal(t, calledErr, errArgument)
	})
}

func TestRealOsExecutor_OpenFile(t *testing.T) {
	t.Run("it uses builtin `osOpenFile`", func(t *testing.T) {
		called := false
		var calledPath string
		var calledFlag int
		var calledPerm os.FileMode
		calledReturnOsFile := &os.File{}
		var calledReturnErr error

		osOpenfile = func(path string, flag int, perm os.FileMode) (*os.File, error) {
			called = true
			calledPath = path
			calledFlag = flag
			calledPerm = perm
			return calledReturnOsFile, calledReturnErr
		}
		defer func() {
			osOpenfile = os.OpenFile
		}()

		osExecutor := &RealOsExecutor{}

		pathArgument := "/tmp/myfile"
		flagArgument := os.O_CREATE
		permArgument := os.FileMode(0444)

		actualOsFile, actualErr := osExecutor.OpenFile(pathArgument, flagArgument, permArgument)
		require.NoError(t, actualErr)

		assert.True(t, called)
		assert.Equal(t, calledReturnOsFile, actualOsFile)
		assert.Equal(t, calledPath, pathArgument)
		assert.Equal(t, calledFlag, flagArgument)
		assert.Equal(t, calledPerm, permArgument)
	})
}

func TestRealOsExecutor_WriteFile(t *testing.T) {
	t.Run("it uses builtin `ioutilWriteFile`", func(t *testing.T) {
		called := false
		var calledPath string
		var calledData []byte
		var calledPerm os.FileMode
		var calledReturnErr error

		ioutilWriteFile = func(path string, data []byte, perm os.FileMode) error {
			called = true
			calledPath = path
			calledData = data
			calledPerm = perm
			return calledReturnErr
		}
		defer func() {
			ioutilWriteFile = ioutil.WriteFile
		}()

		osExecutor := &RealOsExecutor{}

		pathArgument := "/tmp/myfile"
		dataArgument := []byte{1, 2, 3, 4}
		permArgument := os.FileMode(0444)

		actualErr := osExecutor.WriteFile(pathArgument, dataArgument, permArgument)
		require.NoError(t, actualErr)

		assert.True(t, called)
		assert.Equal(t, calledPath, pathArgument)
		assert.Equal(t, calledData, dataArgument)
		assert.Equal(t, calledPerm, permArgument)
	})
}

func TestRealOsExecutor_ExpandTilde(t *testing.T) {
	t.Run(
		"with a 'path' containing a tilde (~), it expands it",
		func(t *testing.T) {
			path := "~/"

			osExecutor := &RealOsExecutor{}
			currentUser, err := osExecutor.CurrentUser()
			require.NoError(t, err)

			expectedPath := fmt.Sprintf("%s/", currentUser.HomeDir)

			actualPath, err := osExecutor.ExpandTilde(path)
			require.NoError(t, err)

			assert.Equal(t, actualPath, expectedPath)
		},
	)

	t.Run(
		"with a 'path' not containing a tilde (~), it does not expand it",
		func(t *testing.T) {
			t.Parallel()

			path := "/home/syndbg"

			osExecutor := &RealOsExecutor{}
			expandedPath, err := osExecutor.ExpandTilde(path)
			require.NoError(t, err)

			assert.Equal(t, expandedPath, path)
		},
	)
}

func TestRealOsExecutor_Getenv(t *testing.T) {
	t.Run(
		"it uses builtin 'osGetenv'",
		func(t *testing.T) {
			keyArg := "HOME"

			called := false
			var calledKey string
			calledReturn := "predefinedbytest"

			osGetenv = func(key string) string {
				called = true
				calledKey = key

				return calledReturn
			}
			defer func() {
				osGetenv = os.Getenv
			}()

			osExecutor := &RealOsExecutor{}
			actualEnv := osExecutor.Getenv(keyArg)

			assert.True(t, called)
			assert.Equal(t, calledKey, keyArg)
			assert.Equal(t, actualEnv, calledReturn)
		},
	)
}

func TestRealOsExecutor_GetOS(t *testing.T) {
	t.Run(
		"it uses `runtime.GOOS`",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}
			actualOS := osExecutor.GetOS()

			assert.Equal(t, actualOS, runtime.GOOS)
		},
	)
}

func TestRealOsExecutor_Remove(t *testing.T) {
	t.Run(
		"it uses builtin 'osRemove'",
		func(t *testing.T) {
			called := false
			var calledNameArg string
			calledReturnError := errors.New("fake")

			osRemove = func(name string) error {
				called = true
				calledNameArg = name
				return calledReturnError

			}
			defer func() {
				osRemove = os.Remove
			}()

			nameArg := "example"

			osExecutor := &RealOsExecutor{}
			err := osExecutor.Remove(nameArg)

			assert.True(t, called)
			assert.Equal(t, calledNameArg, nameArg)
			assert.Equal(t, err, calledReturnError)
		},
	)
}

func TestRealOsExecutor_Create(t *testing.T) {
	t.Run(
		"it uses builtin 'osCreate'",
		func(t *testing.T) {
			called := false
			var calledNameArg string
			calledOsFileInfo := &os.File{}
			calledReturnError := errors.New("fake")

			osCreate = func(name string) (*os.File, error) {
				called = true
				calledNameArg = name
				return calledOsFileInfo, calledReturnError
			}
			defer func() {
				osCreate = os.Create
			}()

			nameArg := "example"

			osExecutor := &RealOsExecutor{}
			actualReturn, actualErr := osExecutor.Create(nameArg)

			assert.True(t, called)
			assert.Equal(t, nameArg, calledNameArg)
			assert.Equal(t, calledReturnError, actualErr)
			assert.Equal(t, calledOsFileInfo, actualReturn)
		},
	)
}

func TestRealOsExecutor_ReadFile(t *testing.T) {
	t.Run(
		"it uses builtin 'ioutilReadFile'",
		func(t *testing.T) {
			called := false
			var calledFilenameArg string
			calledBytes := []byte("test")
			calledReturnError := errors.New("fake")

			ioutilReadFile = func(filename string) (bytes []byte, e error) {
				called = true
				calledFilenameArg = filename
				return calledBytes, calledReturnError
			}
			defer func() {
				ioutilReadFile = ioutil.ReadFile
			}()

			filenameArg := "example"

			osExecutor := &RealOsExecutor{}
			actualReturn, actualErr := osExecutor.ReadFile(filenameArg)

			assert.True(t, called)
			assert.Equal(t, filenameArg, calledFilenameArg)
			assert.Equal(t, calledReturnError, actualErr)
			assert.Equal(t, calledBytes, actualReturn)
		},
	)
}

func TestRealOsExecutor_CurrentUser(t *testing.T) {
	t.Run("it uses builtin `userCurrent`", func(t *testing.T) {
		called := false
		fakeUser := &user.User{}
		calledErr := errors.New("example")

		userCurrent = func() (*user.User, error) {
			called = true
			return fakeUser, calledErr
		}
		defer func() {
			userCurrent = user.Current
		}()

		osExecutor := &RealOsExecutor{}

		actualReturn, actualErr := osExecutor.CurrentUser()
		assert.True(t, called)
		assert.Equal(t, fakeUser, actualReturn)
		assert.Equal(t, calledErr, actualErr)
	})
}

func TestRealOsExecutor_ExecuteWithStreams(t *testing.T) {
	t.Run(
		"with at least env variable specified, "+
			"it runs command with env var set to it and specified, cmd, args, env, dir, stdout and stderr",
		func(t *testing.T) {
			fakeCmd := &exec.Cmd{}

			called := false
			var calledName string
			var calledArgs []string

			execCommand = func(name string, arg ...string) *exec.Cmd {
				called = true
				calledName = name
				calledArgs = arg
				return fakeCmd
			}
			defer func() {
				execCommand = exec.Command
			}()

			osExecutor := &RealOsExecutor{}

			cmdArg := "echo"
			argsArg := []string{"example"}
			envArg := []string{"GOPKGS_EXAMPLE=1"}
			dirArg := "/tmp"
			var stdoutArg, stderrArg bytes.Buffer

			actualErr := osExecutor.ExecuteWithStreams(
				cmdArg,
				argsArg,
				envArg,
				dirArg,
				&stdoutArg,
				&stderrArg,
			)

			assert.True(t, called)
			assert.Equal(t, cmdArg, calledName)
			assert.Equal(t, calledArgs, argsArg)
			assert.Contains(t, actualErr.Error(), "executing command failed")

			assert.Equal(t, envArg, fakeCmd.Env)
			assert.Equal(t, dirArg, fakeCmd.Dir)
			assert.Equal(t, &stdoutArg, fakeCmd.Stdout)
			assert.Equal(t, &stderrArg, fakeCmd.Stderr)
		},
	)

	t.Run(
		"with no env variable specified, it runs command with no env var set",
		func(t *testing.T) {
			fakeCmd := &exec.Cmd{}

			called := false
			var calledName string
			var calledArgs []string

			execCommand = func(name string, arg ...string) *exec.Cmd {
				called = true
				calledName = name
				calledArgs = arg
				return fakeCmd
			}
			defer func() {
				execCommand = exec.Command
			}()

			osExecutor := &RealOsExecutor{}

			cmdArg := "echo"
			argsArg := []string{"example"}
			var envArg []string
			dirArg := "/tmp"
			var stdoutArg, stderrArg bytes.Buffer

			actualErr := osExecutor.ExecuteWithStreams(
				cmdArg,
				argsArg,
				envArg,
				dirArg,
				&stdoutArg,
				&stderrArg,
			)

			assert.True(t, called)
			assert.Equal(t, cmdArg, calledName)
			assert.Equal(t, calledArgs, argsArg)
			assert.Contains(t, actualErr.Error(), "executing command failed")

			assert.Equal(t, envArg, fakeCmd.Env)
			assert.Equal(t, dirArg, fakeCmd.Dir)
			assert.Equal(t, &stdoutArg, fakeCmd.Stdout)
			assert.Equal(t, &stderrArg, fakeCmd.Stderr)
		},
	)
}

func TestRealOsExecutor_Execute(t *testing.T) {
	t.Run(
		"with at least env variable specified, "+
			"it runs command with env var set to it",
		func(t *testing.T) {
			fakeCmd := &exec.Cmd{}

			called := false
			var calledName string
			var calledArgs []string

			execCommand = func(name string, arg ...string) *exec.Cmd {
				called = true
				calledName = name
				calledArgs = arg
				return fakeCmd
			}
			defer func() {
				execCommand = exec.Command
			}()

			osExecutor := &RealOsExecutor{}

			cmdArg := "echo"
			argsArg := []string{"example"}
			envArg := []string{"GOPKGS_EXAMPLE=1"}
			dirArg := "/tmp"

			actualStdout, actualStderr, actualErr := osExecutor.Execute(
				cmdArg,
				argsArg,
				envArg,
				dirArg,
			)

			assert.True(t, called)
			assert.Equal(t, cmdArg, calledName)
			assert.Equal(t, calledArgs, argsArg)
			assert.Contains(t, actualErr.Error(), "executing command failed")
			// NOTE: Nothing written, that's why stdout and stderr are nil
			assert.Nil(t, actualStdout)
			assert.Nil(t, actualStderr)

			assert.Equal(t, envArg, fakeCmd.Env)
			assert.Equal(t, dirArg, fakeCmd.Dir)
		},
	)

	t.Run(
		"with no env variable specified, "+
			"it runs command with no env var set and specified, cmd, args, env, dir, stdout and stderr",
		func(t *testing.T) {
			fakeCmd := &exec.Cmd{}

			called := false
			var calledName string
			var calledArgs []string

			execCommand = func(name string, arg ...string) *exec.Cmd {
				called = true
				calledName = name
				calledArgs = arg
				return fakeCmd
			}
			defer func() {
				execCommand = exec.Command
			}()

			osExecutor := &RealOsExecutor{}

			cmdArg := "echo"
			argsArg := []string{"example"}
			var envArg []string
			dirArg := "/tmp"

			actualStdout, actualStderr, actualErr := osExecutor.Execute(
				cmdArg,
				argsArg,
				envArg,
				dirArg,
			)

			assert.True(t, called)
			assert.Equal(t, cmdArg, calledName)
			assert.Equal(t, calledArgs, argsArg)
			assert.Contains(t, actualErr.Error(), "executing command failed")
			// NOTE: Nothing written, that's why stdout and stderr are nil
			assert.Nil(t, actualStdout)
			assert.Nil(t, actualStderr)

			assert.Equal(t, envArg, fakeCmd.Env)
			assert.Equal(t, dirArg, fakeCmd.Dir)
		},
	)
}

func TestRealOsExecutor_RemoveAll(t *testing.T) {
	t.Run("it uses builtin `osRemoveAll`", func(t *testing.T) {
		called := false
		var calledPathArgument string
		var calledErr error

		osRemoveAll = func(path string) error {
			called = true
			calledPathArgument = path
			return calledErr
		}
		defer func() {
			osRemoveAll = os.RemoveAll
		}()

		pathArgument := "/tmp/example"
		osExecutor := &RealOsExecutor{}

		err := osExecutor.RemoveAll(pathArgument)
		require.NoError(t, err)

		assert.True(t, called)
		assert.Equal(t, calledPathArgument, pathArgument)
	})
}

func TestRealOsExecutor_TempDir(t *testing.T) {
	t.Run("it uses builtin `ioutilTempDir`", func(t *testing.T) {
		called := false
		var calledDir string
		var calledPrefix string
		var fakeString string
		var fakeErr error

		ioutilTempDir = func(dir, prefix string) (string, error) {
			called = true
			calledDir = dir
			calledPrefix = prefix
			return fakeString, fakeErr
		}
		defer func() {
			ioutilTempDir = ioutil.TempDir
		}()

		osExecutor := &RealOsExecutor{}

		dirArg := "/tmp/example"
		prefixArg := "pattern"

		actualReturn, actualErr := osExecutor.TempDir(dirArg, prefixArg)
		require.True(t, called)

		assert.Equal(t, calledDir, dirArg)
		assert.Equal(t, calledPrefix, prefixArg)

		assert.Equal(t, fakeString, actualReturn)
		assert.Equal(t, fakeErr, actualErr)
	})
}

func TestRealOsExecutor_TempFile(t *testing.T) {
	t.Run("it uses builtin `ioutilTempFile`", func(t *testing.T) {
		called := false
		var calledDir string
		var calledPattern string
		var fakeFile *os.File
		var fakeErr error

		ioutilTempFile = func(dir, pattern string) (*os.File, error) {
			called = true
			calledDir = dir
			calledPattern = pattern
			return fakeFile, fakeErr
		}
		defer func() {
			ioutilTempFile = ioutil.TempFile
		}()

		osExecutor := &RealOsExecutor{}

		dirArg := "/tmp/example"
		patternArg := "pattern"

		actualReturn, actualErr := osExecutor.TempFile(dirArg, patternArg)
		require.True(t, called)

		assert.Equal(t, calledDir, dirArg)
		assert.Equal(t, calledPattern, patternArg)

		assert.Equal(t, fakeFile, actualReturn)
		assert.Equal(t, fakeErr, actualErr)
	})
}

func TestRealOsExecutor_ReadDir(t *testing.T) {
	t.Run("it uses builtin `ioutilReadDir`", func(t *testing.T) {
		called := false
		var calledDirname string
		var fakeInfos []os.FileInfo
		var fakeErr error

		ioutilReadDir = func(dirname string) ([]os.FileInfo, error) {
			called = true
			calledDirname = dirname
			return fakeInfos, fakeErr
		}
		defer func() {
			ioutilReadDir = ioutil.ReadDir
		}()

		osExecutor := &RealOsExecutor{}

		dirnameArg := "/tmp/example"

		actualReturn, actualErr := osExecutor.ReadDir(dirnameArg)
		require.True(t, called)

		assert.Equal(t, calledDirname, dirnameArg)

		assert.Equal(t, fakeInfos, actualReturn)
		assert.Equal(t, fakeErr, actualErr)
	})
}
