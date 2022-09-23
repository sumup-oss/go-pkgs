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
	"io"
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

	"github.com/sumup-oss/go-pkgs/template"

	"github.com/palantir/stacktrace"
)

const (
	DefaultFakeGitServerContainerName = "fake_git_server"
	GitDockerImage                    = "winlu/docker-git-server"
	DefaultExposedGitPort             = 2222
	DefaultExposedGitBindHost         = "127.0.0.1"
	DefaultGitDockerDaemonHost        = "unix:///var/run/docker.sock"
	DefaultGitDockerDaemonCertPath    = ""
	DefaultGitDockerDaemonTLSVerify   = "1"
)

type FakeGitServer struct {
	Host                string
	Port                int
	dockerContainerName string
	dockerEnv           []string
}

func NewFakeGitServer(
	host string,
	port int,
	dockerContainerName string,
	dockerDaemonHost string,
	dockerCertPath string,
	dockerTLSVerify string,
) *FakeGitServer {
	dockerEnv := []string{
		fmt.Sprintf("DOCKER_HOST=%s", dockerDaemonHost),
		fmt.Sprintf("DOCKER_CERT_PATH=%s", dockerCertPath),
		fmt.Sprintf("DOCKER_TLS_VERIFY=%s", dockerTLSVerify),
	}

	return &FakeGitServer{
		Host:                host,
		Port:                port,
		dockerContainerName: dockerContainerName,
		dockerEnv:           dockerEnv,
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

	dockerDaemonHost := os.Getenv("TEST_GIT_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultGitDockerDaemonHost
	}

	dockerDaemonCertPath := os.Getenv("TEST_GIT_DOCKER_CERT_PATH")
	if dockerDaemonCertPath == "" {
		dockerDaemonCertPath = DefaultGitDockerDaemonCertPath
	}

	dockerDaemonTLSVerify := os.Getenv("TEST_GIT_DOCKER_TLS_VERIFY")
	if dockerDaemonTLSVerify == "" {
		dockerDaemonTLSVerify = DefaultGitDockerDaemonTLSVerify
	}

	dockerContainerName := os.Getenv("TEST_GIT_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		dockerContainerName = DefaultFakeGitServerContainerName
	}

	return NewFakeGitServer(gitHost, gitPortInt, dockerContainerName, dockerDaemonHost, dockerDaemonCertPath, dockerDaemonTLSVerify)
}

func DefaultGitServer() *FakeGitServer {
	return NewFakeGitServer(
		DefaultExposedGitBindHost,
		DefaultExposedGitPort,
		DefaultFakeGitServerContainerName,
		DefaultGitDockerDaemonHost,
		DefaultGitDockerDaemonCertPath,
		DefaultGitDockerDaemonTLSVerify,
	)
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
		s.dockerContainerName,
		"sh",
		"add_git_user.sh",
		username,
		strings.TrimSuffix(string(authorizedKey), "\n"),
	}
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv
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
	// nolint:gosec
	cmd := exec.Command("git", "clone", gitEnv["GIT_REPOSITORY_URL"], gitEnv["GIT_WORK_DIR"])
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = GitEnvToArray(gitEnv)
	err := cmd.Run()

	return gitEnv["GIT_WORK_DIR"], err
}

func (s *FakeGitServer) copyFile(srcPath, dstPath string) error {
	srcfd, err := os.Open(srcPath)
	if err != nil {
		return stacktrace.Propagate(err, "failed to open file %s", dstPath)
	}
	defer srcfd.Close()

	dstfd, err := os.Create(dstPath)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create file %s", dstPath)
	}
	defer dstfd.Close()

	_, err = io.Copy(dstfd, srcfd)
	if err != nil {
		return stacktrace.Propagate(err, "failed to copy file contents from %s to %s", srcPath, dstPath)
	}

	err = os.Chmod(dstPath, 0644)

	return stacktrace.Propagate(err, "failed to change file permissions, path %s", dstPath)
}

func (s *FakeGitServer) renderTemplate(srcPath, dstPath string, templateData interface{}) error {
	content, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return stacktrace.Propagate(err, "failed reading template file %s", srcPath)
	}
	tpl := template.MustParseTemplate(srcPath, string(content))

	dstPath = strings.TrimSuffix(dstPath, ".gotmpl")
	dstfd, err := os.Create(dstPath)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create file %s", dstPath)
	}
	defer dstfd.Close()

	err = tpl.Execute(dstfd, templateData)

	return stacktrace.Propagate(err, "failed to execute template")
}

func (s *FakeGitServer) AddDirToGitRepo(
	gitEnv map[string]string,
	dir string,
	templateData interface{},
) error {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		relativePath := strings.TrimPrefix(path, dir)
		relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
		dstPath := filepath.Join(gitEnv["GIT_WORK_DIR"], relativePath)

		if info.IsDir() {
			err = os.MkdirAll(dstPath, 0755)

			return stacktrace.Propagate(err, "failed to create directory %s", dstPath)
		}

		if filepath.Ext(info.Name()) == ".gotmpl" && templateData != nil {
			err = s.renderTemplate(path, dstPath, templateData)

			return stacktrace.Propagate(err, "failed to render template %s", path)
		}

		err = s.copyFile(path, dstPath)

		return stacktrace.Propagate(err, "failed to copy file %s", path)
	})

	if err != nil {
		return stacktrace.Propagate(err, "failed to copy directory %s", dir)
	}

	gitEnvArray := GitEnvToArray(gitEnv)

	//nolint:gosec
	cmd := exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err = cmd.Run()
	if err != nil {
		return stacktrace.Propagate(err, "failed to git add files")
	}

	//nolint:gosec
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

	//nolint:gosec
	cmd = exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err = cmd.Run()

	return stacktrace.Propagate(err, "failed to git push files")
}

func (s *FakeGitServer) AddFilesToGitRepo(gitEnv map[string]string, files map[string]string) error {
	for relativeFilePath, fileContent := range files {
		fullFilePath := filepath.Join(gitEnv["GIT_WORK_DIR"], relativeFilePath)
		dir := filepath.Dir(fullFilePath)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return stacktrace.Propagate(err, "failed to create dir")
		}

		err = ioutil.WriteFile(fullFilePath, []byte(fileContent), 0755) // nolint: gosec
		if err != nil {
			return stacktrace.Propagate(err, "failed to write file")
		}
	}

	gitEnvArray := GitEnvToArray(gitEnv)

	//nolint:gosec
	cmd := exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "add", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err := cmd.Run()
	if err != nil {
		return stacktrace.Propagate(err, "failed to git add files")
	}

	//nolint:gosec
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

	//nolint:gosec
	cmd = exec.Command("git", "-C", gitEnv["GIT_WORK_DIR"], "push")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = gitEnvArray
	err = cmd.Run()

	return stacktrace.Propagate(err, "failed to git push files")
}

// nolint: thelper
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

	if len(gitRepoFiles) > 0 {
		err = s.AddFilesToGitRepo(gitEnv, gitRepoFiles)
		require.Nil(t, err)
	}

	return gitEnv, gitWorkDir
}

func (s *FakeGitServer) PrepareGitServerWithArgs() error {
	//nolint:errcheck
	s.StopGitServer()

	//nolint:gosec
	cmd := exec.Command("docker", "run",
		"-p",
		fmt.Sprintf("%s:%d:22/tcp", s.Host, s.Port),
		fmt.Sprintf("--name=%s", s.dockerContainerName),
		"-d",
		"--rm",
		"-ti",
		GitDockerImage,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv
	err := cmd.Run()
	if err != nil {
		return err
	}

	// NOTE: Check 15 times with interval 1 second,
	// for healthiness of previously started git server.
	isHealthy := false
	log.Printf("Waiting for GIT %s to be healthy\n", GitDockerImage)

	for i := 0; i < 15; i++ {
		if s.IsGitHealthy() {
			isHealthy = true

			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return fmt.Errorf("GIT %s still not healthy after 15 attempts", GitDockerImage)
	}

	log.Printf("GIT %s is healthy\n", GitDockerImage)

	return nil
}

func (s *FakeGitServer) StopGitServer() error {
	// NOTE: Ignore error since we clean optimistically
	//nolint:gosec
	cmd := exec.Command("docker", "rm", "-fv", s.dockerContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv

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
