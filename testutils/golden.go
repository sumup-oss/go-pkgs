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

package testutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertGolden(t *testing.T, path string, actual []byte) {
	t.Helper()

	// nolint:goconst
	if os.Getenv("UPDATE_GOLDEN") == "on" {
		err := ioutil.WriteFile(path, actual, 0644) // nolint: gosec
		require.Nil(t, err)
	}

	expected, err := ioutil.ReadFile(path)
	require.Nil(t, err)

	assert.Equal(t, string(expected), string(actual))
}

func AssertGoldenTemplate(t *testing.T, path string, templateVars map[string]string, actual []byte) {
	t.Helper()

	if os.Getenv("UPDATE_GOLDEN") == "on" {
		templateContent := actual
		for k, v := range templateVars {
			re := regexp.MustCompile(fmt.Sprintf("(?m)%s", regexp.QuoteMeta(v)))
			templateContent = re.ReplaceAll(
				templateContent,
				[]byte(fmt.Sprintf("__GOLDENVAR_%s", k)),
			)
		}

		err := ioutil.WriteFile(path, templateContent, 0644) // nolint: gosec
		require.Nil(t, err)
	}

	expected, err := ioutil.ReadFile(path)
	require.Nil(t, err)

	for k, v := range templateVars {
		re := regexp.MustCompile("(?m)" + regexp.QuoteMeta(fmt.Sprintf("__GOLDENVAR_%s", k)))
		expected = re.ReplaceAll(expected, []byte(v))
	}

	assert.Equal(t, string(expected), string(actual))
}
