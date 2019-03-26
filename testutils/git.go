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
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	"github.com/palantir/stacktrace"
)

const (
	DockerTestContainerName   = "beacon_test_container"
	DockerImage               = "winlu/docker-git-server"
	DefaultExposedGitPort     = 2222
	DefaultExposedGitBindHost = "127.0.0.1"
)

type FakeGitServer struct {
	Host string
	Port int
}

func NewFakeGitServer(host string, port int) *FakeGitServer {
	return &FakeGitServer{
		Host: host,
		Port: port,
	}
}

func NewFakeGitServerFromEnv() *FakeGitServer {
	gitHost := os.Getenv("TEST_GIT_HOST")
	if gitHost == "" {
		gitHost = DefaultExposedGitBindHost
	}

	var gitPortInt int
	gitPort := os.Getenv("TEST_GIT_PORT")
	if gitPort == "" {
		gitPortInt = DefaultExposedGitPort
	} else {
		gitPortParsed, err := strconv.ParseInt(gitPort, 10, 32)
		if err != nil {
			panic(err)
		}

		gitPortInt = int(gitPortParsed)
	}

	return NewFakeGitServer(gitHost, gitPortInt)
}

func DefaultGitServer() *FakeGitServer {
	return NewFakeGitServer(DefaultExposedGitBindHost, DefaultExposedGitPort)
}

func (s *FakeGitServer) PrepareGitEnv(privKeyPath, username, repositoryName, gitWorkDir string) map[string]string {
	return map[string]string{
		"GIT_SSH_COMMAND": fmt.Sprintf(
			"ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s",
			privKeyPath,
		),
		"GIT_REPOSITORY_URL": fmt.Sprintf(
			"ssh://%s@%s:%d/~/%s.git",
			username,
			s.Host,
			s.Port,
			repositoryName,
		),
		"GIT_WORK_DIR":    gitWorkDir,
		"SSH_AUTH_SOCK":   "",
		"GIT_SSH_VARIANT": "ssh",
	}
}

func GitEnvToArray(gitEnv map[string]string) []string {
	res := make([]string, 0)
	for k, v := range gitEnv {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}
	return res
}

func (s *FakeGitServer) AddRemoteGitUser(username string, privKey *rsa.PrivateKey) error {
	pub, err := ssh.NewPublicKey(&privKey.PublicKey)
	if err != nil {
		return err
	}

	authorizedKey := ssh.MarshalAuthorizedKey(pub)
	args := []string{
		"exec",
		DockerTestContainerName,
		"sh",
		"add_git_user.sh",
		username,
		strings.TrimSuffix(string(authorizedKey), "\n"),
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// NOTE: docker exec requires a pty (or at least something that behaves like such).
	err = cmd.Start()
	if err != nil {
		return stacktrace.Propagate(err, "failed adding git user with ssh key")
	}
	err = cmd.Wait()
	return stacktrace.Propagate(err, "failed adding git user with ssh key")
}

func (s *FakeGitServer) CreateRemoteGitRepo(privKeyPath, username, repositoryName string) error {
	args := []string{
		"-i",
		privKeyPath,
		"-o StrictHostKeyChecking=no",
		"-o UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", username, s.Host),
		"-p",
		strconv.FormatInt(int64(s.Port), 10),
		"init",
		fmt.Sprintf("%s.git", repositoryName),
	}
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{`SSH_AUTH_SOCK=""`}
	err := cmd.Run()
	return stacktrace.Propagate(
		err,
		"failed to initialize `%s`'s `%s.git` in GIT server",
		username,
		repositoryName,
	)
}

func (s *FakeGitServer) CloneRemoteGitRepo(gitEnv map[string]string) (string, error) {
	cmd := exec.Command("git", []string{"clone", gitEnv["GIT_REPOSITORY_URL"], gitEnv["GIT_WORK_DIR"]}...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = GitEnvToArray(gitEnv)
	err := cmd.Run()
	return gitEnv["GIT_WORK_DIR"], err
}

func (s *FakeGitServer) AddFilesToGitRepo(gitEnv map[string]string, files map[string]string) error {
	for relativeFilePath, fileContent := range files {
		fullFilePath := filepath.Join(gitEnv["GIT_WORK_DIR"], relativeFilePath)
		dir := filepath.Dir(fullFilePath)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return stacktrace.Propagate(err, "failed to create dir")
		}

		err = ioutil.WriteFile(fullFilePath, []byte(fileContent), 0755)
		if err != nil {
			return stacktrace.Propagate(err, "failed to write file")
		}
	}

	gitEnvArray := GitEnvToArray(gitEnv)

	cmd := exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err := cmd.Run()
	if err != nil {
		return stacktrace.Propagate(err, "failed to git add files")
	}

	cmd = exec.Command(
		"git",
		"-C",
		gitEnv["GIT_WORK_DIR"],
		"-c",
		"user.name=AddFilesToGitRepo",
		"-c",
		"user.email=example@example.com",
		"commit",
		"-m",
		"Add from `AddFilesToGitRepo`",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err = cmd.Run()
	if err != nil {
		return stacktrace.Propagate(err, "failed to git commit files")
	}

	cmd = exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err = cmd.Run()
	return stacktrace.Propagate(err, "failed to git push files")
}

func (s *FakeGitServer) SetupForGitRepo(
	t *testing.T,
	tempDir string,
	gitRepoFiles map[string]string,
) (map[string]string, string) {
	// NOTE: Also used as remote user name
	repositoryName := RandString(16)
	username := repositoryName
	privKeyPath, privkey := GenerateAndWritePrivateKey(
		t,
		tempDir,
		"id_rsa",
	)
	gitWorkDir := filepath.Join(tempDir, "gitRepo")
	gitEnv := s.PrepareGitEnv(privKeyPath, username, repositoryName, gitWorkDir)

	err := s.AddRemoteGitUser(username, privkey)
	require.Nil(t, err)

	err = s.CreateRemoteGitRepo(privKeyPath, username, repositoryName)
	require.Nil(t, err)

	_, err = s.CloneRemoteGitRepo(gitEnv)
	require.Nil(t, err)

	err = s.AddFilesToGitRepo(gitEnv, gitRepoFiles)
	require.Nil(t, err)

	return gitEnv, gitWorkDir
}

func (s *FakeGitServer) PrepareGitServerWithArgs() error {
	//nolint:errcheck
	s.StopGitServer()

	cmd := exec.Command("docker", "run",
		"-p",
		fmt.Sprintf("%s:%d:22/tcp", s.Host, s.Port),
		fmt.Sprintf("--name=%s", DockerTestContainerName),
		"-d",
		"--rm",
		"-ti",
		DockerImage,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	// NOTE: Check 15 times with interval 1 second,
	// for healthiness of previously started git server.
	isHealthy := false
	log.Printf("Waiting for GIT %s to be healthy\n", DockerImage)

	for i := 0; i < 15; i++ {
		if s.IsGitHealthy() {
			isHealthy = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return fmt.Errorf("GIT %s still not healthy after 15 attempts", DockerImage)
	}

	log.Printf("GIT %s is healthy\n", DockerImage)
	return nil
}

func (s *FakeGitServer) StopGitServer() error {
	// NOTE: Ignore error since we clean optimistically
	cmd := exec.Command("docker", "rm", "-fv", DockerTestContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *FakeGitServer) IsGitHealthy() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
