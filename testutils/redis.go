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
	RedisDockerImage             = "redis"
	DefaultRedisVersion          = "5"
	DefaultRedisDockerDaemonHost = "unix:///var/run/docker.sock"
	DefaultRedisHost             = "127.0.0.1"
	DefaultRedisPort             = 6379
)

type FakeRedisServer struct {
	Host                string
	Port                int
	version             string
	dockerContainerName string
	dockerEnv           []string
}

func NewFakeRedis5Server(
	dockerContainerName,
	dockerDaemonHost,
	host string,
	port int,
) *FakeRedisServer {
	return NewFakeRedisServer(
		DefaultRedisVersion,
		dockerContainerName,
		dockerDaemonHost,
		host,
		port,
	)
}

func NewFakeRedis5ServerFromEnv() *FakeRedisServer {
	host := os.Getenv("TEST_REDIS_HOST")
	if host == "" {
		host = DefaultRedisHost
	}

	var portInt int
	port := os.Getenv("TEST_REDIS_PORT")
	if port == "" {
		portInt = DefaultRedisPort
	} else {
		portParsed, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}

		portInt = int(portParsed)
	}

	dockerDaemonHost := os.Getenv("TEST_REDIS_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultRedisDockerDaemonHost
	}

	dockerContainerName := os.Getenv("TEST_REDIS_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		panic(`Blank "TEST_REDIS_DOCKER_CONTAINER_NAME"`)
	}

	return NewFakeRedis5Server(
		dockerContainerName,
		dockerDaemonHost,
		host,
		portInt,
	)
}

func NewFakeRedisServerFromEnv() *FakeRedisServer {
	host := os.Getenv("TEST_REDIS_HOST")
	if host == "" {
		host = DefaultRedisHost
	}

	var portInt int
	port := os.Getenv("TEST_REDIS_PORT")
	if port == "" {
		portInt = DefaultRedisPort
	} else {
		portParsed, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}

		portInt = int(portParsed)
	}

	dockerDaemonHost := os.Getenv("TEST_REDIS_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultRedisDockerDaemonHost
	}

	dockerContainerName := os.Getenv("TEST_REDIS_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		panic(`Blank "TEST_REDIS_DOCKER_CONTAINER_NAME"`)
	}

	version := os.Getenv("TEST_REDIS_VERSION")
	if version == "" {
		version = DefaultRedisVersion
	}

	return NewFakeRedisServer(
		version,
		dockerContainerName,
		dockerDaemonHost,
		host,
		portInt,
	)
}

func NewFakeRedisServer(
	version,
	dockerContainerName,
	dockerDaemonHost,
	host string,
	port int,
) *FakeRedisServer {
	dockerEnv := []string{fmt.Sprintf("DOCKER_HOST=%s", dockerDaemonHost)}

	return &FakeRedisServer{
		Host:                host,
		Port:                port,
		version:             version,
		dockerContainerName: dockerContainerName,
		dockerEnv:           dockerEnv,
	}
}

func (s *FakeRedisServer) RunWithArgs() error {
	_ = s.Stop()

	//nolint:gosec
	cmd := exec.Command("docker", "run",
		"-p",
		fmt.Sprintf("%s:%d:6379/tcp", s.Host, s.Port),
		fmt.Sprintf("--name=%s", s.dockerContainerName),
		"-d",
		"--rm",
		"-ti",
		fmt.Sprintf("%s:%s", RedisDockerImage, s.version),
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
	log.Println("Waiting for redis to be healthy")

	for i := 0; i < 15; i++ {
		if s.IsHealthy() {
			isHealthy = true

			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return errors.New("redis is still not healthy after 15 attempts")
	}

	log.Println("redis is healthy")

	return nil
}

func (s *FakeRedisServer) Stop() error {
	// NOTE: Ignore error since we clean optimistically
	// nolint:gosec
	cmd := exec.Command("docker", "rm", "-fv", s.dockerContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv

	return cmd.Run()
}

func (s *FakeRedisServer) IsHealthy() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
