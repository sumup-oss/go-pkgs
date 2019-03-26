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
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
)

// RunCommandInSameProcess is a test helper function to test CLI in integration tests
// by running in the same process and being able to mock/stub behavior.
func RunCommandInSameProcess(cmd *cobra.Command, args []string, output io.Writer) (*cobra.Command, error) {

	// NOTE: When https://github.com/spf13/cobra/pull/822 is merged
	// use separate stdout and stderr. Also allow stdin
	cmd.SetArgs(args)
	cmd.SetOutput(output)
	cmdReturn, err := cmd.ExecuteC()

	return cmdReturn, err
}

// RunCommandInSubprocess is a test helper function to test CLI in end-to-end tests
// by running them in a separate subprocess, but not being able able to mock/stub behavior.
func RunCommandInSubprocess(t *testing.T, args ...string) *exec.Cmd {
	cmdArgs := []string{
		fmt.Sprintf("-test.run=%s", t.Name()),
		"--",
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(os.Args[0], cmdArgs...)
	cmd.Env = []string{
		"GO_WANT_HELPER_PROCESS=1",
	}
	return cmd
}
