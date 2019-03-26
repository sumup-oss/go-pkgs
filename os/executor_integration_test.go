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
	"github.com/mattes/go-expand-tilde"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup-oss/go-pkgs/testutils"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRealOsExecutor_ResolvePath(t *testing.T) {
	t.Run(
		"with path containing `~`, it returns expanded absolute path",
		func(t *testing.T) {
			osExecutor := &RealOsExecutor{}

			pathArg := filepath.Join("~", "mypath")

			actualReturn, actualErr := osExecutor.ResolvePath(pathArg)
			require.Nil(t, actualErr)

			expectedPath, err := tilde.Expand(pathArg)
			require.Nil(t, err)

			assert.Equal(t, expectedPath, actualReturn)
		},
	)

	t.Run(
		"with path not containing `~`, it returns absolute path",
		func(t *testing.T) {

			if runtime.GOOS == "windows" {
				t.Skip("Not supported OS")
			}

			osExecutor := &RealOsExecutor{}

			pathArg := "/home/example/mypath"

			actualReturn, actualErr := osExecutor.ResolvePath(pathArg)
			require.Nil(t, actualErr)

			assert.Equal(t, "/home/example/mypath", actualReturn)
		},
	)

	t.Run(
		"with path not containing `~`, it returns absolute path",
		func(t *testing.T) {

			if runtime.GOOS != "windows" {
				t.Skip("Not supported OS")
			}

			osExecutor := &RealOsExecutor{}

			pathArg := filepath.Join(`C:\\`, "example", "mypath")

			actualReturn, actualErr := osExecutor.ResolvePath(pathArg)
			require.Nil(t, actualErr)

			assert.Equal(t, filepath.Join(`C:\\`, "example", "mypath"), actualReturn)
		},
	)
}

func TestRealOsExecutor_IsFile(t *testing.T) {
	t.Run(
		"with path that is a dir, it returns error",
		func(t *testing.T) {
			osExecutor := &RealOsExecutor{}

			testDir := testutils.TestCwd(t, "os-executor")
			pathArg := filepath.Join(testDir, "dir")

			err := osExecutor.Mkdir(pathArg, 755)
			require.Nil(t, err)

			actualErr := osExecutor.IsFile(pathArg)
			require.NotNil(t, actualErr)
			assert.Equal(t, "not a file", actualErr.Error())
		},
	)

	t.Run(
		"with path that is a file, it returns nil",
		func(t *testing.T) {
			osExecutor := &RealOsExecutor{}

			testDir := testutils.TestCwd(t, "os-executor")
			pathArg := filepath.Join(testDir, "dir")

			err := osExecutor.WriteFile(pathArg, []byte("1234"), 755)
			require.Nil(t, err)

			actualErr := osExecutor.IsFile(pathArg)
			require.Nil(t, actualErr)
		},
	)
}

func TestRealOsExecutor_IsDir(t *testing.T) {
	t.Run(
		"with path that is a file it returns error",
		func(t *testing.T) {
			osExecutor := &RealOsExecutor{}

			testDir := testutils.TestCwd(t, "os-executor")
			pathArg := filepath.Join(testDir, "dir")

			err := osExecutor.WriteFile(pathArg, []byte("1234"), 755)
			require.Nil(t, err)

			actualErr := osExecutor.IsDir(pathArg)
			require.NotNil(t, actualErr)
			assert.Equal(t, "not a dir", actualErr.Error())
		},
	)

	t.Run(
		"with path that is a dir, it returns nil",
		func(t *testing.T) {
			osExecutor := &RealOsExecutor{}

			testDir := testutils.TestCwd(t, "os-executor")
			pathArg := filepath.Join(testDir, "dir")

			err := osExecutor.Mkdir(pathArg, 0755)
			require.Nil(t, err)

			actualErr := osExecutor.IsDir(pathArg)
			require.Nil(t, actualErr)
		},
	)
}
