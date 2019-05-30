package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/sumup-oss/go-pkgs/os/ostest"
)

func TestKubectl_RolloutStatus(t *testing.T) {
	t.Run(
		"when passing 'timeout' and  'resourceName' argument, "+
			"it calls kubectl execute with provided arguments",
		func(t *testing.T) {
			t.Parallel()
			executor := ostest.NewFakeOsExecutor(t)

			executor.On(
				"Execute",
				"kubectl",
				[]string{"-n", "default", "rollout", "status", "deployment/foo", "--timeout", "5s"},
				[]string(nil),
				"",
			).Return([]byte("output is ignored"), []byte(nil), nil)

			kubectl := NewKubectl(executor, "", "svc.cluster.local").(*Kubectl)

			_ = kubectl.RolloutStatus(time.Second*5, "deployment/foo", "default")

			executor.AssertExpectations(t)
		},
	)
}

func TestKubectl_JobStatus(t *testing.T) {
	t.Run("kubectl stdout", func(t *testing.T) {
		tests := []struct {
			Description    string
			KubectlStdout  string
			ExpectedStatus KubernetesJobStatus
			ExpectError    bool
		}{
			{
				Description: "it returns KubernetesJobStatusActive status",
				KubectlStdout: `
{
    "status": {
        "active": 2,
        "startTime": "2019-02-13T13:57:32Z"
    }
}
`,
				ExpectedStatus: KubernetesJobStatusActive,
			},
			{
				Description:    "it returns KubernetesJobStatusUnkown status on json error",
				KubectlStdout:  `invalid_json`,
				ExpectedStatus: KubernetesJobStatusUnknown,
				ExpectError:    true,
			},
			{
				Description: "it returns KubernetesJobStatusComplete status",
				KubectlStdout: `
{
	"status": {
		"completionTime": "2019-02-13T09:26:47Z",
		"conditions": [
			{
				"lastProbeTime": "2019-02-13T09:26:47Z",
				"lastTransitionTime": "2019-02-13T09:26:47Z",
				"status": "True",
				"type": "Complete"
			}
		],
		"startTime": "2019-02-13T09:26:14Z",
		"succeeded": 2
	}
}
`,
				ExpectedStatus: KubernetesJobStatusComplete,
			},
			{
				Description: "it returns KubernetesJobStatusFailed status",
				KubectlStdout: `
{
	"status": {
		"conditions": [
			{
				"lastProbeTime": "2019-02-13T09:31:30Z",
				"lastTransitionTime": "2019-02-13T09:31:30Z",
				"message": "Job has reached the specified backoff limit",
				"reason": "BackoffLimitExceeded",
				"status": "True",
				"type": "Failed"
			}
		],
		"failed": 1,
		"succeeded": 1,
		"startTime": "2019-02-13T09:29:40Z"
	}
}
`,
				ExpectedStatus: KubernetesJobStatusFailed,
			},
			{
				Description:    "it returns KubernetesJobStatusUnkown status on json error",
				KubectlStdout:  `invalid_json`,
				ExpectedStatus: KubernetesJobStatusUnknown,
				ExpectError:    true,
			},
		}

		for _, tc := range tests {
			test := tc
			t.Run(test.Description, func(t *testing.T) {
				t.Parallel()
				executor := ostest.NewFakeOsExecutor(t)

				executor.On(
					"Execute",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]byte(test.KubectlStdout), []byte{}, nil)

				kubectl := NewKubectl(executor, "", "svc.cluster.local").(*Kubectl)

				status, err := kubectl.JobStatus("foo", "default")
				if test.ExpectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.ExpectedStatus, status)
			})
		}
	})

	t.Run(
		"when passing job name it calls kubectl with the correct arguments",
		func(t *testing.T) {
			t.Parallel()
			executor := ostest.NewFakeOsExecutor(t)

			statusJSON := []byte(` {"status": { "succeeded": 1 }} `)
			executor.On(
				"Execute",
				"kubectl",
				[]string{"-n", "default", "get", "job", "foo", "-o", "json"},
				[]string(nil),
				"",
			).Return(statusJSON, []byte{}, nil)

			kubectl := NewKubectl(executor, "", "svc.cluster.local").(*Kubectl)

			_, _ = kubectl.JobStatus("foo", "default")

			executor.AssertExpectations(t)
		},
	)

	t.Run(
		"it returns KuberenetesJobStatusUnknown when kubectl command fails",
		func(t *testing.T) {
			t.Parallel()
			executor := ostest.NewFakeOsExecutor(t)

			statusJSON := []byte(` {"status": { "succeeded": 1 }} `)
			executor.On(
				"Execute",
				"kubectl",
				mock.Anything,
				mock.Anything,
				mock.Anything,
			).Return(statusJSON, []byte{}, assert.AnError)

			kubectl := NewKubectl(executor, "", "svc.cluster.local").(*Kubectl)

			status, err := kubectl.JobStatus("foo", "default")
			assert.Equal(t, assert.AnError, err)
			assert.Equal(t, KubernetesJobStatusUnknown, status)
		})
}
