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
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"runtime"

	"github.com/palantir/stacktrace"

	"github.com/sumup-oss/go-pkgs/os"
)

var (
	linuxIpRouteRegex  = regexp.MustCompile(`(?m)src\s+\b(?P<ip>(?:\d{1,3}\.){3}\d{1,3})\s+`)
	darwinIpRouteRegex = regexp.MustCompile(`(?m)\s+gateway:\s+\b(?P<ip>(?:\d{1,3}\.){3}\d{1,3})\s+`)
)

type DockerBuildOptions struct {
	Hosts      map[string]string
	BuildArgs  []string
	File       string
	Tag        string
	Target     string
	ContextDir string
}

type DockerNetwork struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
	IPAМ struct {
		Driver string `json:"Driver"`
		Config []struct {
			Subnet  string `json:"Subnet"`
			Gateway string `json:"Gateway"`
		} `json:"Config"`
	} `json:"IPAM"`
}

type Docker struct {
	binaryPath      string
	commandExecutor os.CommandExecutor
}

func NewDocker(executor os.CommandExecutor) *Docker {
	return &Docker{
		binaryPath:      "docker",
		commandExecutor: executor,
	}
}

func (docker *Docker) Push(ctx context.Context, image string) error {
	args := []string{"push", image}
	stdout, stderr, err := docker.commandExecutor.ExecuteContext(ctx, "docker", args, nil, "")
	return stacktrace.Propagate(err, "Stderr: %s, Stdout: %s", stderr, stdout)
}

func (docker *Docker) Pull(ctx context.Context, image string) error {
	args := []string{"pull", image}
	stdout, stderr, err := docker.commandExecutor.ExecuteContext(ctx, "docker", args, nil, "")
	return stacktrace.Propagate(err, "Stderr: %s, Stdout: %s", stderr, stdout)
}

func (docker *Docker) buildArgs(options *DockerBuildOptions) []string {
	args := []string{"build", "-f", options.File, "--tag", options.Tag}

	if options.Target != "" {
		args = append(args, "--target", options.Target)
	}

	if options.Hosts != nil {
		for name, address := range options.Hosts {
			args = append(args, fmt.Sprintf("--add-host=%s:%s", name, address))
		}
	}

	if options.BuildArgs != nil {
		for _, buildArg := range options.BuildArgs {
			args = append(args, fmt.Sprintf("--build-arg=%s", buildArg))
		}
	}

	args = append(args, options.ContextDir)
	return args
}

func (docker *Docker) Build(ctx context.Context, options *DockerBuildOptions) error {
	args := docker.buildArgs(options)
	stdout, stderr, err := docker.commandExecutor.ExecuteContext(ctx, "docker", args, nil, "")
	return stacktrace.Propagate(err, "Stderr: %s, Stdout: %s", stderr, stdout)
}

func (docker *Docker) Tag(ctx context.Context, oldImage, newImage string) error {
	args := []string{"tag", oldImage, newImage}
	stdout, stderr, err := docker.commandExecutor.ExecuteContext(ctx, "docker", args, nil, "")
	return stacktrace.Propagate(err, "Stderr: %s, Stdout: %s", stderr, stdout)
}

func (docker *Docker) Login(ctx context.Context, username, password, registryUrl string) error {
	args := []string{"login", "-u", username, "-p", password, registryUrl}
	stdout, stderr, err := docker.commandExecutor.ExecuteContext(ctx, "docker", args, nil, "")
	return stacktrace.Propagate(err, "Stderr: %s, Stdout: %s", stderr, stdout)
}

func (docker *Docker) NetworkInspect(ctx context.Context, name string) (*DockerNetwork, error) {
	stdout, _, err := docker.commandExecutor.ExecuteContext(
		ctx,
		docker.binaryPath,
		[]string{"network", "inspect", name},
		nil,
		"",
	)

	if err != nil {
		return nil, stacktrace.Propagate(err, "executing `docker network inspect %s` failed", name)
	}

	var network []*DockerNetwork
	err = json.Unmarshal(stdout, &network)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"json decode command `docker network inspect %s` output failed",
			name,
		)
	}

	if len(network) == 0 {
		return nil, stacktrace.NewError("docker network name not found")
	}

	return network[0], nil
}

func (docker *Docker) NetworkGateway(ctx context.Context, name string) (string, error) {
	network, err := docker.NetworkInspect(ctx, name)
	if err != nil {
		return "", stacktrace.NewError("failed to read docker network configuration")
	}

	if len(network.IPAМ.Config) == 0 {
		return "", stacktrace.NewError("no IPAM configuration found for docker network")
	}

	var gatewayIP string
	// NOTE: Docker has some confirmed regression in terms of setting a `Gateway`.
	// There are scenarios where it might have a bridge with an IP which is the gateway of
	// the docker network, but nothing except subnet CIDR set in the configuration.
	if network.IPAМ.Config[0].Gateway == "" {
		ip, _, err := net.ParseCIDR(network.IPAМ.Config[0].Subnet)
		if err != nil {
			return "", stacktrace.Propagate(err, "failed to parse IP CIDR for docker network CIDR")
		}

		ipv4 := ip.To4()
		ipv4[3]++
		gatewayIP = ipv4.String()
	} else {
		gatewayIP = network.IPAМ.Config[0].Gateway
	}

	// NOTE: Prefer network isolation wherever possible by avoiding binding on the docker gateway.
	// To do this, find the route that is the src when attempting to route to the docker network.
	// This is helpful in a docker-outside docker environment such as
	// Jenkins agents that mount `/var/run/docker.sock`.
	// This is equivalent of `ip route get <docker bridge ip>`.
	switch runtime.GOOS {
	case "linux":
		stdout, _, err := docker.commandExecutor.Execute(
			"ip",
			[]string{"route", "get", gatewayIP},
			nil,
			"",
		)
		if err != nil {
			return gatewayIP, nil
		}

		matches := linuxIpRouteRegex.FindAllStringSubmatch(string(stdout), -1)
		if len(matches) < 1 {
			return gatewayIP, nil
		}

		if len(matches[0]) < 2 {
			return gatewayIP, nil
		}

		gatewayIP = matches[0][1]
	case "darwin":
		stdout, _, err := docker.commandExecutor.ExecuteContext(
			ctx,
			"route",
			[]string{"-n", "get", gatewayIP},
			nil,
			"",
		)
		if err != nil {
			return gatewayIP, nil
		}

		matches := darwinIpRouteRegex.FindAllStringSubmatch(string(stdout), -1)
		if len(matches) < 1 {
			return gatewayIP, nil
		}

		if len(matches[0]) < 2 {
			return gatewayIP, nil
		}

		gatewayIP = matches[0][1]
	}

	return gatewayIP, nil
}
