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
	"io/ioutil"
	"os"
	"testing"
)

// TestDir creates a new temporary dir using specified `dirPrefix`
func TestDir(t *testing.T, dirPrefix string) string {
	t.Helper()

	tmp, err := ioutil.TempDir("", dirPrefix)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	return tmp
}

// TestChdir changes current dir to specified `dir`
func TestChdir(t *testing.T, dir string) {
	t.Helper()

	err := os.Chdir(dir)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
}

// TestCwd creates new temporary dir using `TestDir` and changes to it using `TestChdir`.
func TestCwd(t *testing.T, dirPrefix string) string {
	t.Helper()

	tmp := TestDir(t, dirPrefix)
	TestChdir(t, tmp)

	return tmp
}
