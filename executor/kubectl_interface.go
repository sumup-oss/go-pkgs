package executor

import (
	"time"
)

type KubectlInterface interface {
	Apply(manifest string, namespace string) error
	Delete(manifest string) error
	Create(manifest string) error
	ClusterInfo() error
	GetToken() ([]byte, error)
	GetServiceAccountSecret(namespace, name, dataKeyName string) (string, error)
	GetIngressHost(namespace, name string) (string, error)
	GetServices(namespace string) ([]*KubernetesService, error)
	GetService(name, namespace string) (*KubernetesService, error)
	ApplyConfigmap(name, namespace string, data map[string]string) error
	ApplyService(service *KubernetesService) error
	GetServiceFQDN(namespace, serviceName string) (string, error)
	GetServiceMeta(namespace, serviceName, key string) (string, error)
	GetServicePort(namespace, serviceName, portName string) (string, error)
	GetIngresses(namespace string) ([]*KubernetesIngress, error)
	RolloutStatus(timeout time.Duration, resource, namespace string) error
	JobStatus(name, namespace string) (KubernetesJobStatus, error)
	DeleteResource(namespace, resourceType, resourceName string) error
	DeleteAllResources(namespace, resourceType string) error
	DeleteAllResourcesByLabel(namespace string, labels map[string]string) error
}
