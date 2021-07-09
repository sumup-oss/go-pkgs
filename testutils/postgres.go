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
	PostgresDockerImage             = "postgres"
	DefaultPostgresVersion          = "10"
	DefaultPostgresDockerDaemonHost = "unix:///var/run/docker.sock"
	DefaultPostgresHost             = "127.0.0.1"
	DefaultPostgresPort             = 5432
	DefaultPostgresUser             = "postgres"
	DefaultPostgresPassword         = "password"
	DefaultPostgresDatabase         = "postgres"
)

type FakePostgresServer struct {
	host                string
	port                int
	version             string
	dockerContainerName string
	dockerEnv           []string
	user                string
	password            string
	database            string
}

func NewFakePostgres10Server(
	dockerContainerName,
	dockerDaemonHost,
	host string,
	port int,
	user,
	password,
	database string,
) *FakePostgresServer {
	return NewFakePostgresServer(
		DefaultPostgresVersion,
		dockerContainerName,
		dockerDaemonHost,
		host,
		port,
		user,
		password,
		database,
	)
}

func NewFakePostgres10ServerFromEnv() *FakePostgresServer {
	host := os.Getenv("TEST_POSTGRES_HOST")
	if host == "" {
		host = DefaultPostgresHost
	}

	var portInt int
	port := os.Getenv("TEST_POSTGRES_PORT")
	if port == "" {
		portInt = DefaultPostgresPort
	} else {
		portParsed, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}

		portInt = int(portParsed)
	}

	dockerDaemonHost := os.Getenv("TEST_POSTGRES_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultPostgresDockerDaemonHost
	}

	dockerContainerName := os.Getenv("TEST_POSTGRES_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		panic(`Blank "TEST_POSTGRES_DOCKER_CONTAINER_NAME"`)
	}

	user := os.Getenv("TEST_POSTGRES_USER")
	if user == "" {
		user = DefaultPostgresUser
	}

	password := os.Getenv("TEST_POSTGRES_PASSWORD")
	if password == "" {
		password = DefaultPostgresPassword
	}

	database := os.Getenv("TEST_POSTGRES_DATABASE")
	if database == "" {
		database = DefaultPostgresDatabase
	}

	return NewFakePostgres10Server(
		dockerContainerName,
		dockerDaemonHost,
		host,
		portInt,
		user,
		password,
		database,
	)
}

func NewFakePostgresServerFromEnv() *FakePostgresServer {
	host := os.Getenv("TEST_POSTGRES_HOST")
	if host == "" {
		host = DefaultPostgresHost
	}

	var portInt int
	port := os.Getenv("TEST_POSTGRES_PORT")
	if port == "" {
		portInt = DefaultPostgresPort
	} else {
		portParsed, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			panic(err)
		}

		portInt = int(portParsed)
	}

	dockerDaemonHost := os.Getenv("TEST_POSTGRES_DOCKER_HOST")
	if dockerDaemonHost == "" {
		dockerDaemonHost = DefaultPostgresDockerDaemonHost
	}

	dockerContainerName := os.Getenv("TEST_POSTGRES_DOCKER_CONTAINER_NAME")
	if dockerContainerName == "" {
		panic(`Blank "TEST_POSTGRES_DOCKER_CONTAINER_NAME"`)
	}

	postgresVersion := os.Getenv("TEST_POSTGRES_VERSION")
	if postgresVersion == "" {
		postgresVersion = DefaultPostgresVersion
	}

	user := os.Getenv("TEST_POSTGRES_USER")
	// NOTE: Disable false positive lint offense
	if user == "" {
		user = DefaultPostgresUser
	}

	password := os.Getenv("TEST_POSTGRES_PASSWORD")
	if password == "" {
		password = DefaultPostgresPassword
	}

	database := os.Getenv("TEST_POSTGRES_DATABASE")
	if database == "" {
		database = DefaultPostgresDatabase
	}

	return NewFakePostgresServer(
		postgresVersion,
		dockerContainerName,
		dockerDaemonHost,
		host,
		portInt,
		user,
		password,
		database,
	)
}

func NewFakePostgresServer(
	postgresVersion,
	dockerContainerName,
	dockerDaemonHost,
	host string,
	port int,
	user,
	password,
	database string,
) *FakePostgresServer {
	dockerEnv := []string{fmt.Sprintf("DOCKER_HOST=%s", dockerDaemonHost)}

	return &FakePostgresServer{
		host:                host,
		port:                port,
		version:             postgresVersion,
		dockerContainerName: dockerContainerName,
		dockerEnv:           dockerEnv,
		user:                user,
		password:            password,
		database:            database,
	}
}

func (s *FakePostgresServer) GetDSN(connectTimeout int) string {
	return fmt.Sprintf(
		"host=%s port=%d database=%s user=%s password=%s connect_timeout=%d sslmode=disable",
		s.host,
		s.port,
		s.database,
		s.user,
		s.password,
		connectTimeout,
	)
}

func (s *FakePostgresServer) RunWithArgs() error {
	_ = s.Stop()

	//nolint:gosec
	cmd := exec.Command("docker", "run",
		"-p",
		fmt.Sprintf("%s:%d:5432/tcp", s.host, s.port),
		fmt.Sprintf("--name=%s", s.dockerContainerName),
		"-e",
		fmt.Sprintf("POSTGRES_USER=%s", s.user),
		"-e",
		fmt.Sprintf("POSTGRES_PASSWORD=%s", s.password),
		"-e",
		fmt.Sprintf("POSTGRES_DB=%s", s.database),
		"-d",
		"--rm",
		"-ti",
		fmt.Sprintf("%s:%s", PostgresDockerImage, s.version),
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
	log.Println("Waiting for Postgres to be healthy")

	for i := 0; i < 15; i++ {
		if s.IsHealthy() {
			isHealthy = true

			break
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		return errors.New("postgres is still not healthy after 15 attempts")
	}

	log.Println("Postgres is healthy")

	return nil
}

func (s *FakePostgresServer) Stop() error {
	// NOTE: Ignore error since we clean optimistically
	// nolint:gosec
	cmd := exec.Command("docker", "rm", "-fv", s.dockerContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = s.dockerEnv

	return cmd.Run()
}

func (s *FakePostgresServer) IsHealthy() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return false
	}
	defer conn.Close()

	return true
}
