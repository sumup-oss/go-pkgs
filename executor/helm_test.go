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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup-oss/go-pkgs/os/ostest"
)

func TestNewHelm(t *testing.T) {
	t.Run(
		"it creates helm with specified executor and binpath of `helm` and kubectl version of `1.9`",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := ostest.NewFakeOsExecutor(t)

			actual := NewHelm(osExecutor)

			assert.Equal(t, "helm", actual.binPath)
			assert.Equal(t, "1.9", actual.kubeVersion)
			assert.Equal(t, osExecutor, actual.commandExecutor)
		},
	)
}

func TestHelm_GetManifest(t *testing.T) {
	t.Run(
		"when values does not contain a string with commas inside, "+
			"it does not escape the value",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := ostest.NewFakeOsExecutor(t)

			locationArg := "/tmp/example"
			nameArg := "example"
			namespaceArg := "default"
			valuesArg := map[string]string{}
			valuesArg["excludedAttributes"] = "request"
			stringValuesArg := map[string]string{}

			helmInstance := NewHelm(osExecutor)

			fakeStdout := []byte("fakeStdout")
			fakeStderr := []byte("fakeStderr")

			expectedCmdArgs := []string{
				"template",
				"--name",
				nameArg,
				"--kube-version",
				helmInstance.kubeVersion,
				"--namespace",
				namespaceArg,
				"--set",
				"excludedAttributes=request",
				locationArg,
			}
			osExecutor.On(
				"Execute",
				helmInstance.binPath,
				expectedCmdArgs,
				[]string(nil),
				"",
			).Return(fakeStdout, fakeStderr, nil)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			osExecutor.AssertExpectations(t)

			require.Nil(t, actualErr)
			assert.Equal(t, string(fakeStdout), actual)
		},
	)

	t.Run(
		"when values contain a string with commas inside, "+
			"it escapes the commas and quoting the whole string",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := ostest.NewFakeOsExecutor(t)

			locationArg := "/tmp/example"
			nameArg := "example"
			namespaceArg := "default"
			valuesArg := map[string]string{}
			valuesArg["excludedAttributes"] = "request.headers.cookie,request.headers.authorization,request.headers.proxyAuthorization,request.headers.setCookie*,request.headers.x*,response.headers.cookie,response.headers.authorization,response.headers.proxyAuthorization,response.headers.setCookie*,response.headers.x*"

			stringValuesArg := map[string]string{}

			helmInstance := NewHelm(osExecutor)

			fakeStdout := []byte("fakeStdout")
			fakeStderr := []byte("fakeStderr")

			expectedCmdArgs := []string{
				"template",
				"--name",
				nameArg,
				"--kube-version",
				helmInstance.kubeVersion,
				"--namespace",
				namespaceArg,
				"--set",
				`excludedAttributes=request.headers.cookie\,request.headers.authorization\,request.headers.proxyAuthorization\,request.headers.setCookie*\,request.headers.x*\,response.headers.cookie\,response.headers.authorization\,response.headers.proxyAuthorization\,response.headers.setCookie*\,response.headers.x*`,
				locationArg,
			}
			osExecutor.On(
				"Execute",
				helmInstance.binPath,
				expectedCmdArgs,
				[]string(nil),
				"",
			).Return(fakeStdout, fakeStderr, nil)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			osExecutor.AssertExpectations(t)

			require.Nil(t, actualErr)
			assert.Equal(t, string(fakeStdout), actual)
		},
	)

	t.Run(
		"when values are in string values, "+
			"it uses --set-string instead of --set",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := ostest.NewFakeOsExecutor(t)

			locationArg := "/tmp/example"
			nameArg := "example"
			namespaceArg := "default"
			valuesArg := map[string]string{}
			valuesArg["excludedAttributes"] = "request.headers.cookie,request.headers.authorization,request.headers.proxyAuthorization,request.headers.setCookie*,request.headers.x*,response.headers.cookie,response.headers.authorization,response.headers.proxyAuthorization,response.headers.setCookie*,response.headers.x*"

			stringValuesArg := map[string]string{}
			stringValuesArg["someTest"] = "true"

			helmInstance := NewHelm(osExecutor)

			fakeStdout := []byte("fakeStdout")
			fakeStderr := []byte("fakeStderr")

			expectedCmdArgs := []string{
				"template",
				"--name",
				nameArg,
				"--kube-version",
				helmInstance.kubeVersion,
				"--namespace",
				namespaceArg,
				"--set",
				`excludedAttributes=request.headers.cookie\,request.headers.authorization\,request.headers.proxyAuthorization\,request.headers.setCookie*\,request.headers.x*\,response.headers.cookie\,response.headers.authorization\,response.headers.proxyAuthorization\,response.headers.setCookie*\,response.headers.x*`,
				"--set-string",
				`someTest=true`,
				locationArg,
			}
			osExecutor.On(
				"Execute",
				helmInstance.binPath,
				expectedCmdArgs,
				[]string(nil),
				"",
			).Return(fakeStdout, fakeStderr, nil)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			osExecutor.AssertExpectations(t)

			require.Nil(t, actualErr)
			assert.Equal(t, string(fakeStdout), actual)
		},
	)
}
