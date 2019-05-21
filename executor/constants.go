package executor

const (
	KubernetesJobStatusUnknown KubernetesJobStatus = iota
	KubernetesJobStatusActive
	KubernetesJobStatusComplete
	KubernetesJobStatusFailed
	// kubernetes condition statuses
	kubernetesConditionStatusTrue = "True"
	// kubernetes condition types
	kubernetesJobConditionComplete = "Complete"
	kubernetesJobConditionFailed   = "Failed"
)
