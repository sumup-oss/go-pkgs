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
	"fmt"
	"path"
	"path/filepath"
	"strings"

	stdOs "os"

	"github.com/sumup-oss/go-pkgs/os"
)

type Git struct {
	binPath         string
	dir             string
	url             string
	env             []string
	commandExecutor os.CommandExecutor
}

func NewGit(executor os.CommandExecutor, url, dir string, env []string) *Git {
	var gitEnv []string

	if len(env) < 1 {
		gitEnv = make([]string, 0)
	} else {
		gitEnv = env
	}

	return &Git{
		binPath:         "git",
		dir:             dir,
		url:             url,
		commandExecutor: executor,
		env:             gitEnv,
	}
}

func (git *Git) GetURL() string {
	return git.url
}

func (git *Git) Clone() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"clone", git.url, git.dir},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) CloneWithoutCheckout() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"clone", "--no-checkout", git.url, git.dir},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) CreateAndSwitchToBranch(ctx context.Context, name string) error {
	_, stderr, err := git.commandExecutor.ExecuteContext(
		ctx,
		git.binPath,
		[]string{"-C", git.dir, "checkout", "-b", name},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) SwitchToBranch(ctx context.Context, name string) error {
	_, stderr, err := git.commandExecutor.ExecuteContext(
		ctx,
		git.binPath,
		[]string{"-C", git.dir, "checkout", name},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) DeleteLocalBranch(ctx context.Context, name string) error {
	_, stderr, err := git.commandExecutor.ExecuteContext(
		ctx,
		git.binPath,
		[]string{"-C", git.dir, "branch", "-D", name},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) DeleteRemoteBranch(ctx context.Context, name string) error {
	_, stderr, err := git.commandExecutor.ExecuteContext(
		ctx,
		git.binPath,
		[]string{"-C", git.dir, "push", "origin", "--delete", name},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) HasDiff() (bool, error) {
	output, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "status", "--porcelain"},
		git.env,
		"",
	)
	if err != nil {
		return false, fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return len(output) > 0, nil
}

func (git *Git) Add(file string) error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "add", file},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) Commit(message string) error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "commit", "-m", message},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) Push(destination string, isForce bool) error {
	args := []string{"-C", git.dir, "push"}
	if isForce {
		args = append(args, "--force")
	}

	args = append(args, strings.Split(destination, " ")...)

	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		args,
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("failed to execute git command. Err: %s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) Pull() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "pull"},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) Fetch() error {
	// TODO: add --prune-tags when we update to git > 2.17
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "fetch", "--prune"},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) SwitchAndReset(branch string) error {
	// NOTE: First attempt to reset with origin
	// so that we handle cases where local tip of the branch a.k.a HEAD is behind,
	// but the remote HEAD (origin) is ahead.
	// However if we receive a `branch` that's actually a git commit, `origin/<commit>` is invalid,
	// cause git is awesome.
	// Then git reset `commit` will result into exactly the remote HEAD.
	_, _, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "reset", "--hard", path.Join("origin", branch)},
		git.env,
		"",
	)
	if err != nil {
		_, stderr, err := git.commandExecutor.Execute(
			git.binPath,
			[]string{"-C", git.dir, "reset", "--hard", branch},
			git.env,
			"",
		)
		if err != nil {
			return fmt.Errorf("%s. Stderr: %s", err, stderr)
		}
	}

	return nil
}

func (git *Git) ListBranches() ([]string, error) {
	stdout, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "branch", "-a"},
		git.env,
		"",
	)

	if err != nil {
		return nil, fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	branchesOutput := strings.Split(string(stdout), "\n")

	branches := make([]string, 0)

	for _, branch := range branchesOutput {
		branchName := strings.Trim(branch, "*\r\n ")
		if branchName == "" {
			continue
		}
		branches = append(branches, branchName)
	}

	return branches, nil
}

func (git *Git) ListTags() ([]string, error) {
	stdout, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{
			"-C", git.dir,
			"for-each-ref",
			"--sort=taggerdate",
			"--format",
			"%(refname)",
			"refs/tags",
		},
		git.env,
		"",
	)

	if err != nil {
		return nil, fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	tagsOutput := strings.Split(string(stdout), "\n")

	tags := make([]string, 0)

	for _, tag := range tagsOutput {
		tagName := strings.TrimPrefix(strings.Trim(tag, "*\r\n "), "refs/tags/")
		if tagName == "" {
			continue
		}
		tags = append(tags, tagName)
	}

	return tags, nil
}

func (git *Git) CheckoutClean() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "checkout", "."},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) EnableSparseCheckout() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "config", "core.sparseCheckout", "true"},
		git.env,
		"",
	)

	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) DisableSparseCheckout() error {
	_, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "config", "core.sparseCheckout", "false"},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}

func (git *Git) SetSparseCheckoutPaths(paths []string) error {
	sparsePaths := strings.Join(paths, "\n")
	sparseConfigPath := filepath.Join(git.dir, ".git", "info", "sparse-checkout")

	err := stdOs.WriteFile(sparseConfigPath, []byte(sparsePaths), 0755) //nolint: gosec,mnd
	if err != nil {
		return fmt.Errorf("error writing to sparse checkout config file. Error: %s", err.Error())
	}

	return nil
}

func (git *Git) GetCurrentHash() (string, error) {
	stdout, stderr, err := git.commandExecutor.Execute(
		git.binPath,
		[]string{"-C", git.dir, "rev-parse", "HEAD"},
		git.env,
		"",
	)

	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return strings.Trim(string(stdout), "\n\r "), nil
}

// GetCurrentHashForPath hash of last git commit made.
// `path` can be an absolute or relative path.
// This is needed for mono-repositories, to make sure they only change when their content changes.
// However merge commits will not be the latest commit, since they include no content.
// The commit before the merge commit for the path will be the last.
func (git *Git) GetCurrentHashForPath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	args := []string{
		"-C",
		git.dir,
		"log",
		"-n1",
		"--oneline",
		"--no-abbrev-commit",
		"--",
		path,
	}

	stdout, stderr, err := git.commandExecutor.Execute(git.binPath, args, git.env, "")
	if err != nil {
		return "", fmt.Errorf("failed to execute git command. Err: %s. Stderr: %s", err, stderr)
	}

	stdoutParts := strings.Split(string(stdout), " ")
	if len(stdoutParts) == 0 {
		return "", fmt.Errorf("failed to parse git log output %s. ", stdout)
	}

	return strings.Trim(stdoutParts[0], "\n\r "), nil
}

func (git *Git) Config(ctx context.Context, key, value string) error {
	_, stderr, err := git.commandExecutor.ExecuteContext(
		ctx,
		git.binPath,
		[]string{"-C", git.dir, "config", key, value},
		git.env,
		"",
	)
	if err != nil {
		return fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return nil
}
