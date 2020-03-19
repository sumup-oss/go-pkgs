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
	"github.com/elliotchance/orderedmap"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup-oss/go-pkgs/os"
)

func hasHelmBinary(executor os.OsExecutor) error {
	_, _, err := executor.Execute("helm", nil, nil, "")
	if err != nil {
		return err
	}

	// HACK: Since we only support helm ~ 2.9.x
	_, _, err = executor.Execute("helm --version | grep 2.9", nil, nil, "")
	return err
}

func TestHelm_GetManifest_Integration(t *testing.T) {
	t.Run(
		"when values contain a string with commas inside, "+
			"it escapes the commas and quoting the whole string",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &os.RealOsExecutor{}
			if err := hasHelmBinary(osExecutor); err != nil {
				t.Skipf("No `helm` binary found in $PATH. Error: %s\n", err)
			}

			locationArg := filepath.Join("testdata", "examplechart")
			nameArg := "dashboard-backend"
			namespaceArg := "default"
			valuesArg := orderedmap.NewOrderedMap()
			stringValuesArg := orderedmap.NewOrderedMap()
			stringValuesArg.Set(
				"newRelic.excludedAttributes",
				"request.headers.cookie,request.headers.authorization,request.headers.proxyAuthorization,request.headers.setCookie*,request.headers.x*,response.headers.cookie,response.headers.authorization,response.headers.proxyAuthorization,response.headers.setCookie*,response.headers.x*",
			)

			helmInstance := NewHelm(osExecutor)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			require.Nil(t, actualErr)

			assert.Contains(t, actual, "")
			assert.Contains(
				t,
				actual,
				`value: "request.headers.cookie,request.headers.authorization,request.headers.proxyAuthorization,request.headers.setCookie*,request.headers.x*,response.headers.cookie,response.headers.authorization,response.headers.proxyAuthorization,response.headers.setCookie*,response.headers.x*"`,
			)
		},
	)

	t.Run(
		"when values does not contain a string with commas inside, "+
			"it does not escape the value",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &os.RealOsExecutor{}
			if err := hasHelmBinary(osExecutor); err != nil {
				t.Skipf("No `helm` binary found in $PATH. Error: %s\n", err)
			}

			locationArg := filepath.Join("testdata", "examplechart")
			nameArg := "dashboard-backend"
			namespaceArg := "default"
			valuesArg := orderedmap.NewOrderedMap()
			valuesArg.Set("newRelic.excludedAttributes", "request")
			stringValuesArg := orderedmap.NewOrderedMap()

			helmInstance := NewHelm(osExecutor)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			require.Nil(t, actualErr)

			assert.Contains(t, actual, "")
			assert.Contains(
				t,
				actual,
				`value: "request"`,
			)
		},
	)

	t.Run(
		"when values contains a string with commas, starting with `{` but not closed, "+
			"it returns error that value is not set",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &os.RealOsExecutor{}
			if err := hasHelmBinary(osExecutor); err != nil {
				t.Skipf("No `helm` binary found in $PATH. Error: %s\n", err)
			}

			locationArg := filepath.Join("testdata", "examplechart")
			nameArg := "dashboard-backend"
			namespaceArg := "default"
			valuesArg := orderedmap.NewOrderedMap()
			valuesArg.Set(
				"newRelic.excludedAttributes",
				`{request.headers.cookie\,request.headers.authorization\,request.headers.proxyAuthorization\,request.headers.setCookie*\,request.headers.x*\,response.headers.cookie\,response.headers.authorization\,response.headers.proxyAuthorization\,response.headers.setCookie*\,response.headers.x*`,
			)

			stringValuesArg := orderedmap.NewOrderedMap()

			helmInstance := NewHelm(osExecutor)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			assert.Equal(t, "", actual)

			assert.Contains(
				t,
				actualErr.Error(),
				`failed parsing --set data: key map "newRelic" has no value`,
			)
		},
	)

	t.Run(
		"when values contains a string with commas, inside an object (`{}`), "+
			"it is escaped and parsed successfully",
		func(t *testing.T) {
			t.Parallel()

			osExecutor := &os.RealOsExecutor{}
			if err := hasHelmBinary(osExecutor); err != nil {
				t.Skipf("No `helm` binary found in $PATH. Error: %s\n", err)
			}

			locationArg := filepath.Join("testdata", "examplechart")
			nameArg := "dashboard-backend"
			namespaceArg := "default"
			valuesArg := orderedmap.NewOrderedMap()
			stringValuesArg := orderedmap.NewOrderedMap()
			stringValuesArg.Set(
				"newRelic.excludedAttributes",
				`{request.headers.cookie\,request.headers.authorization\,request.headers.proxyAuthorization\,request.headers.setCookie*\,request.headers.x*\,response.headers.cookie\,response.headers.authorization\,response.headers.proxyAuthorization\,response.headers.setCookie*\,response.headers.x*}`,
			)

			helmInstance := NewHelm(osExecutor)

			actual, actualErr := helmInstance.GetManifest(
				locationArg,
				nameArg,
				namespaceArg,
				valuesArg,
				stringValuesArg,
			)
			assert.Nil(t, actualErr)
			assert.Contains(
				t,
				actual,
				`[request.headers.cookie\\ request.headers.authorization\\ request.headers.proxyAuthorization\\ request.headers.setCookie*\\ request.headers.x*\\ response.headers.cookie\\ response.headers.authorization\\ response.headers.proxyAuthorization\\ response.headers.setCookie*\\ response.headers.x*]`,
			)
		},
	)
}
