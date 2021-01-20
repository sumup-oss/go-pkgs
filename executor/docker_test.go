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

package executor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumup-oss/go-pkgs/os/ostest"
)

func TestNewDocker(t *testing.T) {
	executorArg := &ostest.FakeOsExecutor{}

	actual := NewDocker(executorArg)
	assert.Equal(t, "docker", actual.binaryPath)
	assert.Equal(t, executorArg, actual.commandExecutor)
}

func TestDocker_Push(t *testing.T) {
	t.Run("when pushing does not fail, it returns nil", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		imageArg := "example"

		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"push", imageArg},
			[]string(nil),
			"",
		).Return([]byte{}, []byte{}, nil)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Push(context.Background(), imageArg)
		require.Nil(t, actual)
	})

	t.Run("when pushing fails, it returns error", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		imageArg := "example"

		fakeError := errors.New("fake error")
		fakeStdout := []byte("fake stdout")
		fakeStderr := []byte("fake stderr")
		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"push", imageArg},
			[]string(nil),
			"",
		).Return(fakeStdout, fakeStderr, fakeError)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Push(context.Background(), imageArg)
		require.NotNil(t, actual)
		assert.Contains(t, actual.Error(), fakeError.Error())
		assert.Contains(t, actual.Error(), string(fakeStdout))
		assert.Contains(t, actual.Error(), string(fakeStderr))
	})
}

func TestDocker_Pull(t *testing.T) {
	t.Run("when pulling does not fail, it returns nil", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		imageArg := "example"

		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"pull", imageArg},
			[]string(nil),
			"",
		).Return([]byte{}, []byte{}, nil)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Pull(context.Background(), imageArg)
		require.Nil(t, actual)
	})

	t.Run("when pulling fails, it returns error", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		imageArg := "example"

		fakeError := errors.New("fake error")
		fakeStdout := []byte("fake stdout")
		fakeStderr := []byte("fake stderr")
		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"pull", imageArg},
			[]string(nil),
			"",
		).Return(fakeStdout, fakeStderr, fakeError)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Pull(context.Background(), imageArg)
		require.NotNil(t, actual)
		assert.Contains(t, actual.Error(), fakeError.Error())
		assert.Contains(t, actual.Error(), string(fakeStdout))
		assert.Contains(t, actual.Error(), string(fakeStderr))
	})
}

func TestDocker_Tag(t *testing.T) {
	t.Run("when tagging does not fail, it returns nil", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		oldImageArg := "example"
		newImageArg := "newexample"

		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"tag", oldImageArg, newImageArg},
			[]string(nil),
			"",
		).Return([]byte{}, []byte{}, nil)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Tag(context.Background(), oldImageArg, newImageArg)
		require.Nil(t, actual)
	})

	t.Run("when tagging fails, it returns error", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		oldImageArg := "example"
		newImageArg := "newexample"

		fakeError := errors.New("fake error")
		fakeStdout := []byte("fake stdout")
		fakeStderr := []byte("fake stderr")
		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"tag", oldImageArg, newImageArg},
			[]string(nil),
			"",
		).Return(fakeStdout, fakeStderr, fakeError)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Tag(context.Background(), oldImageArg, newImageArg)
		require.NotNil(t, actual)
		assert.Contains(t, actual.Error(), fakeError.Error())
		assert.Contains(t, actual.Error(), string(fakeStdout))
		assert.Contains(t, actual.Error(), string(fakeStderr))
	})
}

func TestDocker_Build(t *testing.T) {
	// NOTE: While all the possibilities of code execution in terms of combinatorics are much more than this,
	// to save time writing tests, just test the fewest possible scenarios that cover the conditional logic.
	t.Run(
		"when buildings fails, it returns error",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				ContextDir: ".",
				Tag:        "mytag",
			}

			fakeError := errors.New("fake error")
			fakeStdout := []byte("fake stdout")
			fakeStderr := []byte("fake stderr")
			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return(fakeStdout, fakeStderr, fakeError)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.NotNil(t, actual)
			assert.Contains(t, actual.Error(), fakeError.Error())
			assert.Contains(t, actual.Error(), string(fakeStdout))
			assert.Contains(t, actual.Error(), string(fakeStderr))
		},
	)

	t.Run(
		"when buildings succeeds and `options.Target` is blank, "+
			"it docker builds specified file up and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				Target:     "",
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.Target` is present, "+
			"it docker builds specified file up to specified target and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				Target:     "mytarget",
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--target",
					optionsArg.Target,
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.Target` is present, "+
			"it docker builds specified file up to specified target and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				Target:     "mytarget",
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--target",
					optionsArg.Target,
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.Hosts` have at least one host, "+
			"it docker builds specified file, add hosts and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				Hosts:      map[string]string{"examplehost": "exampleaddress"},
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--add-host=examplehost:exampleaddress",
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.BuildArgs` have at least one build argument, "+
			"it docker builds specified file, adds build args and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				BuildArgs:  []string{"EXAMPLE=VALUE"},
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--build-arg=EXAMPLE=VALUE",
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.CacheFrom` have exactly one cache from image, "+
			"it docker builds specified file, adds the cache from image and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				CacheFrom:  []string{"example-image:tag"},
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--cache-from",
					optionsArg.CacheFrom[0],
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

	t.Run(
		"when buildings succeeds and `options.CacheFrom` have more than one cache from image, "+
			"it docker builds specified file, adds the cache from image and tags to specified tag",
		func(t *testing.T) {
			executorArg := &ostest.FakeOsExecutor{}

			optionsArg := &DockerBuildOptions{
				File:       "./examplefile",
				CacheFrom:  []string{"example-image:tag1", "example-image:tag2"},
				ContextDir: ".",
				Tag:        "mytag",
			}

			executorArg.On(
				"ExecuteContext",
				context.Background(),
				"docker",
				[]string{
					"build",
					"-f",
					optionsArg.File,
					"--tag",
					optionsArg.Tag,
					"--cache-from",
					optionsArg.CacheFrom[0],
					"--cache-from",
					optionsArg.CacheFrom[1],
					optionsArg.ContextDir,
				},
				[]string(nil),
				"",
			).Return([]byte{}, []byte{}, nil)

			dockerInstance := NewDocker(executorArg)
			actual := dockerInstance.Build(context.Background(), optionsArg)
			require.Nil(t, actual)
		},
	)

}

func TestDocker_Login(t *testing.T) {
	t.Run("when log-in does not fail, it returns nil", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		usernameArg := "example"
		passwordArg := "examplePass"
		registryUrlArg := "exampleRegistry"

		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"login", "-u", usernameArg, "-p", passwordArg, registryUrlArg},
			[]string(nil),
			"",
		).Return([]byte{}, []byte{}, nil)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Login(context.Background(), usernameArg, passwordArg, registryUrlArg)
		require.Nil(t, actual)
	})

	t.Run("when log-in fails, it returns error", func(t *testing.T) {
		executorArg := &ostest.FakeOsExecutor{}
		usernameArg := "example"
		passwordArg := "examplePass"
		registryUrlArg := "exampleRegistry"

		fakeError := errors.New("fake error")
		fakeStdout := []byte("fake stdout")
		fakeStderr := []byte("fake stderr")
		executorArg.On(
			"ExecuteContext",
			context.Background(),
			"docker",
			[]string{"login", "-u", usernameArg, "-p", passwordArg, registryUrlArg},
			[]string(nil),
			"",
		).Return(fakeStdout, fakeStderr, fakeError)

		dockerInstance := NewDocker(executorArg)
		actual := dockerInstance.Login(context.Background(), usernameArg, passwordArg, registryUrlArg)
		require.NotNil(t, actual)
		assert.Contains(t, actual.Error(), fakeError.Error())
		assert.Contains(t, actual.Error(), string(fakeStdout))
		assert.Contains(t, actual.Error(), string(fakeStderr))
	})
}
