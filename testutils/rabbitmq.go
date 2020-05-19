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
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	log "github.com/sumup-oss/go-pkgs/logger"
)

const (
	RabbitMqDockerImage             = "rabbitmq"
	DefaultRabbitMqVersion          = "3"
	DefaultRabbitMqDockerDaemonHost = "unix:///var/run/docker.sock"
	DefaultRabbitMqHost             = "127.0.0.1"
	DefaultRabbitMqPort             = 5672
	DefaultRabbitMqUser             = "guest"
	DefaultRabbitMqPassword         = "guest"
)

type FakeRabbitMq struct {
	host                string
	port                int
	version             string
	dockerContainerName string
	dockerEnv           []string
	user                string
	password            string
}

func NewFakeRabbitMqFromEnv() *FakeRabbitMq {
	host := os.Getenv("TEST_RABBITMQ_HOST")
	if host == "" {
		host = DefaultRabbitMqHost
	}

	dockerDaemonHost := os.Getenv("TEST_RABBITMQ_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultRabbitMqDockerDaemonHost
	}

	var portInt int
	port := os.Getenv("TEST_RABBITMQ_PORT")
	if port == "" {
		portInt = DefaultRabbitMqPort
	} else {
		portParsed, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}

		portInt = int(portParsed)
	}

	dockerContainerName := os.Getenv("TEST_RABBITMQ_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		panic(`Blank "TEST_RABBITMQ_DOCKER_CONTAINER_NAME"`)
	}

	rabbitmqVersion := os.Getenv("TEST_RABBITMQ_VERSION")
	if rabbitmqVersion == "" {
		rabbitmqVersion = DefaultRabbitMqVersion
	}

	user := os.Getenv("TEST_RABBITMQ_USER")
	// NOTE: Disable false positive lint offense
	if user == "" {
		user = DefaultRabbitMqUser
	}

	password := os.Getenv("TEST_RABBITMQ_PASSWORD")
	if password == "" {
		password = DefaultRabbitMqPassword
	}

	return NewFakeRabbitMq(
		rabbitmqVersion,
		dockerContainerName,
		dockerDaemonHost,
		host,
		portInt,
		user,
		password,
	)
}

func (s *FakeRabbitMq) GetConnectionString() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		s.user,
		s.password,
		s.host,
		s.port,
	)
}

func NewFakeRabbitMq(
	rabbitmqVersion,
	dockerContainerName,
	dockerDaemonHost,
	host string,
	port int,
	user,
	password string,
) *FakeRabbitMq {
	dockerEnv := []string{fmt.Sprintf("DOCKER_HOST=%s", dockerDaemonHost)}

	return &FakeRabbitMq{
		host:                host,
		port:                port,
		version:             rabbitmqVersion,
		dockerContainerName: dockerContainerName,
		dockerEnv:           dockerEnv,
		user:                user,
		password:            password,
	}
}

func (s *FakeRabbitMq) RunWithArgs() error {
	_ = s.Stop()

	//nolint:gosec
	cmd := exec.Command("docker", "run",
		"-p",
		fmt.Sprintf("%s:%d:5672/tcp", s.host, s.port),
		fmt.Sprintf("--name=%s", s.dockerContainerName),
		"-e",
		fmt.Sprintf("RABBITMQ_DEFAULT_USER=%s", s.user),
		"-e",
		fmt.Sprintf("RABBITMQ_DEFAULT_PASS=%s", s.password),
		"-d",
		"--rm",
		"-ti",
		fmt.Sprintf("%s:%s", RabbitMqDockerImage, s.version),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv
	err := cmd.Run()
	if err != nil {
		return err
	}

	// NOTE: Check 15 times with interval 1 second,
	// for healthiness of previously started server.
	isHealthy := false
	log.Println("Waiting for RabbitMq to be healthy")

	for i := 0; i < 15; i++ {
		if s.IsHealthy() {
			isHealthy = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return errors.New("rabbitmq is still not healthy after 15 attempts")
	}

	log.Println("Rabbitmq is healthy")
	return nil
}

func (s *FakeRabbitMq) Stop() error {
	// NOTE: Ignore error since we clean optimistically
	// nolint:gosec
	cmd := exec.Command("docker", "rm", "-fv", s.dockerContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv
	return cmd.Run()
}

func (s *FakeRabbitMq) IsHealthy() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
