package kubernetes

import (
	"fmt"

	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/palantir/stacktrace"
)

type ClientSet struct {
	clientSet                    *kubernetes.Clientset
	kubeConfig                   *rest.Config
	inCluster                    bool
	kubernetesServiceAccountName string
	kubernetesNamespace          string
}

func NewKubernetesClientSet(
	kubeconfigPath,
	kubernetesServiceAccountName,
	kubernetesNamespace string,
	inCluster bool,
) (*ClientSet, error) {
	var kubeconfig *rest.Config
	var err error

	if inCluster {
		kubeconfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get in-cluster kubeconfig")
		}

	} else {
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to parse kubeconfig")
		}
	}

	clientSet, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"failed to create kubernetes client set from kubeconfig",
		)
	}

	return &ClientSet{
		inCluster:                    inCluster,
		clientSet:                    clientSet,
		kubeConfig:                   kubeconfig,
		kubernetesNamespace:          kubernetesNamespace,
		kubernetesServiceAccountName: kubernetesServiceAccountName,
	}, nil
}

func (c *ClientSet) GetJWT() (string, error) {
	if c.inCluster {
		return c.kubeConfig.BearerToken, nil
	}

	serviceAccountInfo, err := c.
		clientSet.
		CoreV1().
		ServiceAccounts(c.kubernetesNamespace).
		Get(c.kubernetesServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"failed to get kubernetes dashboard service account %s from namespace %s",
			c.kubernetesServiceAccountName,
			c.kubernetesNamespace,
		)
	}

	var secretName string

	for _, serviceAccountSecret := range serviceAccountInfo.Secrets {
		if strings.Contains(serviceAccountSecret.Name, "token") {
			secretName = serviceAccountSecret.Name
			break
		}
	}

	if len(secretName) == 0 {
		return "", fmt.Errorf(
			"no kubernetes service account secret token from namespace %s for service account: %s",
			c.kubernetesNamespace,
			c.kubernetesServiceAccountName,
		)
	}

	secret, err := c.clientSet.CoreV1().Secrets(c.kubernetesNamespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", stacktrace.Propagate(
			err, "failed to get kubernetes service account secret token data from namespace "+
				"%s for service account: %s",
			c.kubernetesNamespace,
			c.kubernetesServiceAccountName,
		)
	}

	token := secret.Data["token"]
	if token == nil {
		return "", stacktrace.Propagate(
			err,
			"no token data found in kubernetes service account secret token data "+
				"from namespace %s for service account: %s",
			c.kubernetesNamespace,
			c.kubernetesServiceAccountName,
		)
	}

	return string(token), nil
}

func (c *ClientSet) CoreV1() corev1.CoreV1Interface {
	return c.clientSet.CoreV1()
}
