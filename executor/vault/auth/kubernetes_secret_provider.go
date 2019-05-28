// Copyright 2018 SumUp Ltd.
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

package auth

import (
	"fmt"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/palantir/stacktrace"
)

const (
	DefaultKubernetesAuthBackendName = "kubernetes"
)

// KubernetesSecretProvider provides a vault secret by issuing a auth/kubernetes/login.
//
// https://www.vaultproject.io/docs/auth/kubernetes.html
type KubernetesSecretProvider struct {
	vaultRoleID                    string
	kubernetesAuthBackendLoginPath string
	kubernetesClientSet            KubernetesClientSet
}

// NewKubernetesSecretProvider creates KubernetesSecretProvider instance.
func NewKubernetesSecretProvider(
	vaultRoleID,
	vaultKubernetesAuthBackendName string,
	kubernetesClientSet KubernetesClientSet,
) *KubernetesSecretProvider {
	return &KubernetesSecretProvider{
		vaultRoleID: vaultRoleID,
		kubernetesAuthBackendLoginPath: fmt.Sprintf(
			"auth/%s/login",
			vaultKubernetesAuthBackendName,
		),
		kubernetesClientSet: kubernetesClientSet,
	}
}

// GetSecret logins using the provided vault client and returns the returned vault secret.
func (a *KubernetesSecretProvider) GetSecret(client Client) (*vaultapi.Secret, error) {
	jwt, err := a.kubernetesClientSet.GetJWT()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get service account's JWT")
	}

	if len(jwt) == 0 {
		return nil, stacktrace.NewError("empty JWT service account JWT token")
	}

	authData := map[string]interface{}{
		"role": a.vaultRoleID,
		"jwt":  jwt,
	}

	secret, err := client.RawWrite(a.kubernetesAuthBackendLoginPath, authData)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"failed to login in Vault with role %s",
			a.vaultRoleID,
		)
	}

	return secret, nil
}
