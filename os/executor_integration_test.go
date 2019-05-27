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

func TestRealOsExecutor_CopyFile(t *testing.T) {
	t.Run(
		"when src file is not existing, it fails and return error",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			srcArg := "/tmp/NOTexistingEXAmpleCopyFile"
			err := osExecutor.CopyFile(srcArg, "/tmp/example1234Notexisting")
			require.Error(t, err, "failed to open file: %s", srcArg)
		},
	)

	t.Run(
		"when src file is existing, but dst is also existing, it overwrites the dst and writes src",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			srcFd, err := osExecutor.TempFile("", "")
			require.Nil(t, err, "failed to create temporary file")

			srcArg := srcFd.Name()
			err = osExecutor.WriteFile(srcArg, []byte("SRC_EXAMPLE"), 0755)
			require.Nil(t, err, "failed to write temporary file")

			dstFd, err := osExecutor.TempFile("", "")
			require.Nil(t, err, "failed to create temporary file")

			dstArg := dstFd.Name()
			err = osExecutor.WriteFile(dstArg, []byte("DST_EXAMPLE"), 0755)
			require.Nil(t, err, "failed to write temporary file")


			err = osExecutor.CopyFile(srcArg, dstArg)
			assert.Nil(t, err)

			dstContents, err := osExecutor.ReadFile(dstArg)
			require.Nil(t, err)
			assert.Equal(t, "SRC_EXAMPLE", string(dstContents))
		},
	)
}

func TestRealOsExecutor_CopyDir(t *testing.T) {
	t.Run(
		"when src dir is not existing, it fails and return error",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			srcArg := "/tmp/notExistingDir1TestOsExecutor"
			err := osExecutor.CopyDir(srcArg, "/tmp/example1234Notexisting")
			require.Error(t, err, "failed to stat file: %s", srcArg)
		},
	)

	t.Run(
		"when src dir is existing, but dst is also existing, it overwrites the dst and writes src",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			srcArg, err := osExecutor.TempDir("", "")
			require.Nil(t, err, "failed to create temporary dir")

			srcExampleFile := filepath.Join(srcArg, "example.txt")
			err = osExecutor.WriteFile(srcExampleFile, []byte("SRC_EXAMPLE"), 0755)
			require.Nil(t, err, "failed to write temporary file")

			dstArg, err := osExecutor.TempDir("", "")
			require.Nil(t, err, "failed to create temporary dir")

			dstExampleFile := filepath.Join(dstArg, "example.txt")
			err = osExecutor.WriteFile(dstExampleFile, []byte("DST_EXAMPLE"), 0755)
			require.Nil(t, err, "failed to write temporary file")


			err = osExecutor.CopyDir(srcArg, dstArg)
			assert.Nil(t, err)

			dstContents, err := osExecutor.ReadFile(dstExampleFile)
			require.Nil(t, err)
			assert.Equal(t, "SRC_EXAMPLE", string(dstContents))
		},
	)
}


func TestRealOsExecutor_RemoveContents(t *testing.T) {
	t.Run(
		"when dir is empty, it does not delete the dir",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			pathArg, err := osExecutor.TempDir("", "")
			require.Nil(t, err, "failed to create temporary dir")

			err = osExecutor.RemoveContents(pathArg, -1)
			require.Nil(t, err)

			_, err = osExecutor.Stat(pathArg)
			require.Nil(t, err)
		},
	)

	t.Run(
		"when dir contains nested files and dirs, it does not delete the dir and deletes all files and" +
			" dirs inside",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &RealOsExecutor{}

			pathArg, err := osExecutor.TempDir("", "")
			require.Nil(t, err, "failed to create temporary dir")

			// NOTE: Setup nested dir structure
			f1Path := filepath.Join(pathArg, "f1")
			err = osExecutor.WriteFile(f1Path, []byte("1"), 0755)
			require.Nil(t, err)

			f2Path := filepath.Join(pathArg, "f2")
			err = osExecutor.WriteFile(f2Path, []byte("2"), 0755)
			require.Nil(t, err)

			dir1Path := filepath.Join(pathArg, "dir1")
			err = osExecutor.Mkdir(dir1Path, 0755)
			require.Nil(t, err)

			nested1Path := filepath.Join(dir1Path, "nested1")
			err = osExecutor.WriteFile(nested1Path, []byte("n1"), 0755)
			require.Nil(t, err)

			err = osExecutor.RemoveContents(pathArg, -1)
			require.Nil(t, err)

			_, err = osExecutor.Stat(pathArg)
			require.Nil(t, err)

			for _, path := range []string{f1Path, f2Path, dir1Path, nested1Path} {
				_, err := osExecutor.Stat(path)
				require.True(t, osExecutor.IsNotExist(err))
			}
		},
	)
}