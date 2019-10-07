package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	pkgOs "github.com/sumup-oss/go-pkgs/os"
)

type (
	KubectlServiceAccountInfo struct {
		Secrets []struct {
			Name string `json:"name"`
		} `json:"secrets"`
	}
	KubectlIngressInfo struct {
		Spec struct {
			Rules []struct {
				Host string `json:"host"`
			} `json:"rules"`
		} `json:"spec"`
	}
	KubectlSecretInfo struct {
		Data struct {
			CaCrt     string `json:"ca.crt"`
			Namespace string `json:"namespace"`
			Token     string `json:"token"`
		} `json:"data"`
	}
	KubernetesJobStatus    int
	kubernetesJobCondition struct {
		Type   string `json:"type"`
		Status string `json:"status"`
	}
	kubernetesJob struct {
		Status struct {
			Conditions []kubernetesJobCondition `json:"conditions"`
			Active     int                      `json:"active"`
			Succeeded  int                      `json:"succeeded"`
			Failed     int                      `json:"failed"`
		} `json:"status"`
	}

	KubernetesServicesResponse struct {
		Items []*KubernetesService `json:"items"`
	}

	KubernetesService struct {
		APIVersion string                     `json:"apiVersion"`
		Kind       string                     `json:"kind"`
		Metadata   *KubernetesServiceMetadata `json:"metadata"`
		Spec       *KubernetesServiceSpec     `json:"spec"`
	}

	KubernetesServiceMetadata struct {
		Annotations map[string]string `json:"annotations"`
		Name        string            `json:"name"`
		Namespace   string            `json:"namespace"`
	}

	KubernetesServiceSpec struct {
		ClusterIP             string                         `json:"clusterIP"`
		ExternalTrafficPolicy string                         `json:"externalTrafficPolicy"`
		Ports                 []*KubernetesServiceSpecPort   `json:"ports"`
		Selector              *KubernetesServiceSpecSelector `json:"selector"`
		SessionAffinity       string                         `json:"sessionAffinity"`
		Type                  string                         `json:"type"`
	}

	KubernetesServiceSpecSelector struct {
		App string `json:"app"`
	}

	KubernetesServiceSpecPort struct {
		Name       string      `json:"name"`
		NodePort   json.Number `json:"nodePort,Number"`
		Port       json.Number `json:"port,Number"`
		Protocol   string      `json:"protocol"`
		TargetPort json.Number `json:"targetPort,Number"`
	}

	KubernetesIngressesResponse struct {
		Items []*KubernetesIngress `json:"items"`
	}

	KubernetesIngress struct {
		Metadata *KubernetesIngressMetadata `json:"metadata"`
		Spec     *KubernetesIngressSpec     `json:"spec"`
	}

	KubernetesIngressMetadata struct {
		Annotations map[string]string `json:"annotations"`
		Name        string            `json:"name"`
		Namespace   string            `json:"namespace"`
	}

	KubernetesIngressSpec struct {
		Rules []*KubernetesIngressRule `json:"rules"`
	}

	KubernetesIngressRule struct {
		Host string `json:"host"`
	}

	KubernetesIngressFirstAddress struct {
		Host string `json:"host"`
	}

	Kubectl struct {
		commandExecutor          pkgOs.CommandExecutor
		GlobalOptions            map[string]string
		commandString            string
		kubernetesInternalDomain string
	}
)

func NewKubectl(
	commandExecutor pkgOs.CommandExecutor,
	kubectlContext,
	kubernetesInternalDomain string,
) *Kubectl {
	globalOptions := make(map[string]string)

	if len(kubectlContext) > 0 {
		globalOptions["context"] = kubectlContext
	}

	return &Kubectl{
		commandExecutor:          commandExecutor,
		GlobalOptions:            globalOptions,
		commandString:            "kubectl",
		kubernetesInternalDomain: kubernetesInternalDomain,
	}
}

func (k *Kubectl) ResetExecutor(commandExecutor pkgOs.CommandExecutor) pkgOs.CommandExecutor {
	old := k.commandExecutor
	k.commandExecutor = commandExecutor
	return old
}

func (k *Kubectl) compileCommand() []string {
	var options = make([]string, len(k.GlobalOptions)/2)

	for key, value := range k.GlobalOptions {
		options = append(options, fmt.Sprintf("--%s=%s", key, value))
	}

	return options
}

func (k *Kubectl) executeCommand(args []string, env []string) ([]byte, []byte, error) {
	args = append(args, k.compileCommand()...)
	return k.commandExecutor.Execute(k.commandString, args, env, "")
}

func (k *Kubectl) Apply(manifest string, namespace string) error {
	commandArgs := append([]string{"apply"}, "-f", manifest)

	if namespace != "" {
		commandArgs = append(commandArgs, "-n", namespace)
	}

	_, _, err := k.executeCommand(commandArgs, nil)
	return err
}

func (k *Kubectl) Delete(manifest string) error {
	commandArgs := append([]string{"delete", "--force"}, "-f", manifest)
	_, _, err := k.executeCommand(commandArgs, nil)
	return err
}

func (k *Kubectl) Create(manifest string) error {
	commandArgs := append([]string{"create"}, "-f", manifest)
	_, _, err := k.executeCommand(commandArgs, nil)
	return err
}

func (k *Kubectl) ClusterInfo() error {
	_, _, err := k.executeCommand([]string{"cluster-info"}, nil)
	return err
}

func (k *Kubectl) GetToken() ([]byte, error) {
	stdout, _, err := k.executeCommand(
		[]string{"get", "-n", "kube-system", "serviceaccount", "kubernetes-dashboard-admin", "-o", "json"},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var serviceAccountInfo *KubectlServiceAccountInfo

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err = decoder.Decode(&serviceAccountInfo); err != nil {
		return nil, err
	}

	stdout, _, err = k.executeCommand(
		[]string{"get", "-n", "kube-system", "secret", serviceAccountInfo.Secrets[0].Name, "-o", "json"},
		nil,
	)
	if err != nil {
		return nil, err
	}
	var secretInfo *KubectlSecretInfo

	decoder = json.NewDecoder(bytes.NewReader(stdout))
	if err = decoder.Decode(&secretInfo); err != nil {
		return nil, err
	}

	rawToken, err := base64.StdEncoding.DecodeString(secretInfo.Data.Token)
	if err != nil {
		return nil, err
	}

	return rawToken, nil
}

func (k *Kubectl) GetIngressHost(namespace, name string) (string, error) {
	stdout, stderr, err := k.executeCommand(
		[]string{"get", "-n", namespace, "ingress", name, "-o", "json"},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	var kubernetesIngress *KubectlIngressInfo

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err := decoder.Decode(&kubernetesIngress); err != nil {
		return "", err
	}

	if kubernetesIngress.Spec.Rules == nil {
		return "", nil
	}

	if len(kubernetesIngress.Spec.Rules) < 1 {
		return "", nil
	}

	return kubernetesIngress.Spec.Rules[0].Host, nil
}

func (k *Kubectl) GetServiceFQDN(namespace, serviceName string) (string, error) {
	return fmt.Sprintf("%s.%s.%s", serviceName, namespace, k.kubernetesInternalDomain), nil
}

func (k *Kubectl) GetServiceMeta(namespace, serviceName, key string) (string, error) {
	stdout, stderr, err := k.executeCommand(
		[]string{
			"get",
			"-n",
			namespace,
			"service",
			serviceName,
			"-o",
			fmt.Sprintf("jsonpath='{.metadata.%s}'", key),
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return strings.Trim(string(stdout), "' "), nil
}

func (k *Kubectl) GetServicePort(namespace, serviceName, portName string) (string, error) {
	stdout, stderr, err := k.executeCommand(
		[]string{
			"get",
			"-n",
			namespace,
			"service",
			serviceName,
			"-o",
			fmt.Sprintf("jsonpath='{.spec.ports[?(@.name==\"%s\")].port}'", portName),
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	return strings.Trim(string(stdout), "' "), nil
}

func (k *Kubectl) GetServiceAccountSecret(namespace, name, dataKeyName string) (string, error) {
	stdout, stderr, err := k.executeCommand(
		[]string{"get", "-n", namespace, "serviceaccount", name, "-o", "json"},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	var serviceAccountInfo *KubectlServiceAccountInfo

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err = decoder.Decode(&serviceAccountInfo); err != nil {
		return "", err
	}

	stdout, stderr, err = k.executeCommand(
		[]string{"get", "-n", namespace, "secret", serviceAccountInfo.Secrets[0].Name, "-o", "json"},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("%s. Stderr: %s", err, stderr)
	}

	var secretInfo *KubectlSecretInfo

	decoder = json.NewDecoder(bytes.NewReader(stdout))
	if err := decoder.Decode(&secretInfo); err != nil {
		return "", err
	}

	switch dataKeyName {
	case "token":
		rawValue, err := base64.StdEncoding.DecodeString(secretInfo.Data.Token)
		if err != nil {
			return "", err
		}

		return string(rawValue), nil
	case "ca.crt":
		rawValue, err := base64.StdEncoding.DecodeString(secretInfo.Data.CaCrt)
		if err != nil {
			return "", err
		}

		return string(rawValue), nil
	default:
		return "", fmt.Errorf("unknown secret name %s", dataKeyName)
	}
}
func (k *Kubectl) GetServices(namespace string) ([]*KubernetesService, error) {
	stdout, _, err := k.executeCommand(
		[]string{"get", "-n", namespace, "service", "-o", "json"},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var kubernetesServicesResponse *KubernetesServicesResponse

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err := decoder.Decode(&kubernetesServicesResponse); err != nil {
		return nil, err
	}

	return kubernetesServicesResponse.Items, nil
}

func (k *Kubectl) GetService(name, namespace string) (*KubernetesService, error) {
	stdout, _, err := k.executeCommand(
		[]string{"get", "-n", namespace, "service", name, "-o", "json"},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var kubernetesService *KubernetesService

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err := decoder.Decode(&kubernetesService); err != nil {
		return nil, err
	}

	return kubernetesService, nil
}

func (k *Kubectl) GetIngresses(namespace string) ([]*KubernetesIngress, error) {
	stdout, _, err := k.executeCommand(
		[]string{"get", "-n", namespace, "ingress", "-o", "json"},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var kubernetesServicesResponse *KubernetesIngressesResponse

	decoder := json.NewDecoder(bytes.NewReader(stdout))
	if err := decoder.Decode(&kubernetesServicesResponse); err != nil {
		return nil, err
	}

	return kubernetesServicesResponse.Items, nil
}

func (k *Kubectl) ApplyConfigmap(name, namespace string, data map[string]string) error {
	fd, err := ioutil.TempFile("", "kubernetes-configmap.yaml")
	if err != nil {
		return nil
	}

	defer func() {
		fd.Close()
		os.Remove(fd.Name())
	}()

	_, err = fd.WriteString("apiVersion: v1\nkind: ConfigMap\n")
	if err != nil {
		return err
	}

	_, err = fd.WriteString(fmt.Sprintf("metadata:\n  name: %s\n  namespace: %s\n", name, namespace))
	if err != nil {
		return err
	}

	_, err = fd.WriteString("data:\n")
	if err != nil {
		return err
	}

	for key, value := range data {
		_, err = fd.WriteString(fmt.Sprintf("  %s: \"%s\"\n", key, value))
		if err != nil {
			return err
		}
	}

	err = fd.Sync()
	if err != nil {
		return err
	}

	return k.Apply(fd.Name(), "")
}

func (k *Kubectl) ApplyService(service *KubernetesService) error {
	fd, err := ioutil.TempFile("", "kubernetes-service.json")
	if err != nil {
		return nil
	}

	defer func() {
		fd.Close()
		os.Remove(fd.Name())
	}()

	bytes, err := json.Marshal(&service)
	if err != nil {
		return err
	}

	_, err = fd.Write(bytes)
	if err != nil {
		return err
	}

	err = fd.Sync()
	if err != nil {
		return err
	}

	return k.Apply(fd.Name(), "")
}

func (k *Kubectl) RolloutStatus(timeout time.Duration, resource, namespace string) error {
	commandArgs := []string{"-n", namespace, "rollout", "status", resource, "--timeout", timeout.String()}
	_, _, err := k.executeCommand(commandArgs, nil)
	return err
}

func (k *Kubectl) JobStatus(name, namespace string) (KubernetesJobStatus, error) {
	commandArgs := []string{"-n", namespace, "get", "job", name, "-o", "json"}
	stdout, _, err := k.executeCommand(commandArgs, nil)
	if err != nil {
		return KubernetesJobStatusUnknown, err
	}

	var job kubernetesJob

	err = json.Unmarshal(stdout, &job)
	if err != nil {
		return KubernetesJobStatusUnknown, err
	}

	for _, cond := range job.Status.Conditions {
		if cond.Type == kubernetesJobConditionComplete && cond.Status == kubernetesConditionStatusTrue {
			return KubernetesJobStatusComplete, nil
		}

		if cond.Type == kubernetesJobConditionFailed && cond.Status == kubernetesConditionStatusTrue {
			return KubernetesJobStatusFailed, nil
		}
	}

	if job.Status.Active > 0 {
		return KubernetesJobStatusActive, nil
	}

	return KubernetesJobStatusUnknown, nil
}

func (k *Kubectl) DeleteResource(namespace, resourceType, resourceName string) error {
	commandArgs := []string{"-n", namespace, "delete", resourceType, resourceName}
	_, stderr, err := k.executeCommand(commandArgs, nil)
	if err != nil {
		return fmt.Errorf("deleting resource failed, err: %v, stderr: %s", err, stderr)
	}

	return nil
}

func (k *Kubectl) DeleteAllResources(namespace, resourceType string) error {
	commandArgs := []string{"-n", namespace, "delete", "--all", resourceType}
	_, stderr, err := k.executeCommand(commandArgs, nil)
	if err != nil {
		return fmt.Errorf("deleting resources failed, err: %v, stderr: %s", err, stderr)
	}

	return nil
}

func (k *Kubectl) DeleteAllResourcesByLabel(namespace string, labels map[string]string) error {
	// NOTE: Delete all resources and ingress which appears not to be deletable by default
	// ref: https://github.com/kubernetes/kubectl/issues/7
	commandArgs := []string{"-n", namespace, "delete", "all,ing"}

	for k, v := range labels {
		commandArgs = append(commandArgs, "-l", fmt.Sprintf("%s=%s", k, v))
	}

	_, stderr, err := k.executeCommand(commandArgs, nil)
	if err != nil {
		return fmt.Errorf("deleting resources failed, err: %v, stderr: %s", err, stderr)
	}

	return nil
}
