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
	"fmt"
	"strings"

	"github.com/elliotchance/orderedmap"

	"github.com/sumup-oss/go-pkgs/os"
)

type Helm struct {
	binPath         string
	kubeVersion     string
	commandExecutor os.CommandExecutor
}

func NewHelm(executor os.CommandExecutor) *Helm {
	return &Helm{
		binPath:         "helm",
		kubeVersion:     "1.9",
		commandExecutor: executor,
	}
}

func (h *Helm) ResetExecutor(commandExecutor os.CommandExecutor) os.CommandExecutor {
	old := h.commandExecutor
	h.commandExecutor = commandExecutor
	return old
}

// GetManifest returns content of a "helm template" substituted manifest.
// Values and string values are maps of string keys to string values.
func (h *Helm) GetManifest(
	location string,
	name string,
	namespace string,
	values,
	stringValues *orderedmap.OrderedMap,
) (string, error) {
	cmdArgs := []string{
		"template",
		"--name",
		name,
		"--kube-version",
		h.kubeVersion,
		"--namespace",
		namespace,
	}

	for _, key := range values.Keys() {
		value, _ := values.Get(key)
		if value == nil {
			continue
		}

		cmdArgs = append(
			cmdArgs,
			h.prepareSetArgument(key.(string), value.(string), false)...,
		)
	}

	for _, key := range stringValues.Keys() {
		value, _ := stringValues.Get(key)
		if value == nil {
			continue
		}

		cmdArgs = append(
			cmdArgs,
			h.prepareSetArgument(key.(string), value.(string), true)...,
		)
	}

	cmdArgs = append(cmdArgs, location)

	stdout, stderr, err := h.commandExecutor.Execute(
		h.binPath,
		cmdArgs,
		nil,
		"",
	)
	if err != nil {
		return "", fmt.Errorf("%s. STDERR: %s", err, stderr)
	}

	return string(stdout), nil
}

func (h *Helm) prepareSetArgument(key, value string, isString bool) []string {
	setCommand := "--set"

	if isString {
		setCommand = "--set-string"

		// HACK: Workaround strict, yet wrong parsing behavior of Helm parser.
		// ref: https://github.com/helm/helm/issues/1556
		// ref: https://github.com/helm/helm/issues/4406
		if strings.Contains(value, ",") {
			value = strings.ReplaceAll(value, ",", "\\,")
		}
	}

	return []string{
		setCommand,
		fmt.Sprintf("%s=%s", key, value),
	}
}

func (h *Helm) Package(dir, dest string) (string, error) {
	args := []string{"package", dir}
	if dest != "" {
		args = append(args, "--destination", dest)
	}

	stdout, stderr, err := h.commandExecutor.Execute(h.binPath, args, nil, "")
	if err != nil {
		return "", fmt.Errorf("%s. STDERR: %s", err, stderr)
	}

	return string(stdout), nil
}
