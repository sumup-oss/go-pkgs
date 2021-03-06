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
	DefaultApproleAuthBackendName = "approle"
)

// AppRoleSecretProvider provides a vault secret by issuing a auth/approle/login.
//
// https://www.vaultproject.io/docs/auth/approle.html
type AppRoleSecretProvider struct {
	vaultRoleID                 string
	vaultSecretID               string
	approleAuthBackendLoginPath string
}

// NewAppRoleSecretProvider creates AppRoleSecretProvider instance.
func NewAppRoleSecretProvider(vaultRoleID, vaultSecretID, approleAuthBackendName string) *AppRoleSecretProvider {
	return &AppRoleSecretProvider{
		vaultRoleID:   vaultRoleID,
		vaultSecretID: vaultSecretID,
		approleAuthBackendLoginPath: fmt.Sprintf(
			"auth/%s/login",
			approleAuthBackendName,
		),
	}
}

// GetSecret logins using the provided vault client and returns the returned vault secret.
func (a *AppRoleSecretProvider) GetSecret(client Client) (*vaultapi.Secret, error) {
	authData := map[string]interface{}{
		"role_id":   a.vaultRoleID,
		"secret_id": a.vaultSecretID,
	}

	secret, err := client.RawWrite(a.approleAuthBackendLoginPath, authData)

	return secret, stacktrace.Propagate(err, "failed to login via approle to Vault")
}
