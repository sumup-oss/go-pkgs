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

package auth_test

import (
	"errors"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/palantir/stacktrace"
	"github.com/stretchr/testify/assert"

	"github.com/sumup-oss/go-pkgs/executor/vault/auth"
)

func TestKubernetesSecretProvider_GetSecret(t *testing.T) {
	t.Run(
		"when the vault kubernetes login is successful, it returns the created secret and no error",
		func(t *testing.T) {
			authData := map[string]interface{}{
				"role": "fooRole",
				"jwt":  "fooJWT",
			}
			expectedSecret := &vaultapi.Secret{
				RequestID: "fooID",
			}
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("RawWrite", "auth/kubernetes/login", authData).Return(expectedSecret, nil).Once()

			fakeKubernetesClientSet := NewFakeKubernetesClientSet(t)
			fakeKubernetesClientSet.On("GetJWT").Return("fooJWT", nil).Once()

			authenticator := auth.NewKubernetesSecretProvider(
				"fooRole",
				auth.DefaultKubernetesAuthBackendName,
				fakeKubernetesClientSet,
			)

			secret, err := authenticator.GetSecret(fakeClient)

			assert.Nil(t, err)
			assert.Equal(t, expectedSecret, secret)
			fakeClient.AssertExpectations(t)
			fakeKubernetesClientSet.AssertExpectations(t)
		},
	)

	t.Run(
		"when the vault kubernetes login fails, it returns nil secret and an error",
		func(t *testing.T) {
			fakeError := errors.New("fakeError")
			authData := map[string]interface{}{
				"role": "fooRole",
				"jwt":  "fooJWT",
			}
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("RawWrite", "auth/kubernetes/login", authData).Return(nil, fakeError).Once()

			fakeKubernetesClientSet := NewFakeKubernetesClientSet(t)
			fakeKubernetesClientSet.On("GetJWT").Return("fooJWT", nil).Once()

			authenticator := auth.NewKubernetesSecretProvider(
				"fooRole",
				auth.DefaultKubernetesAuthBackendName,
				fakeKubernetesClientSet,
			)

			secret, err := authenticator.GetSecret(fakeClient)

			assert.Equal(t, fakeError, stacktrace.RootCause(err))
			assert.Nil(t, secret)
			fakeClient.AssertExpectations(t)
			fakeKubernetesClientSet.AssertExpectations(t)
		},
	)

	t.Run(
		"when the kubernetes client failed to return JWT token, it returns nil secret and an error",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)

			fakeError := errors.New("fakeError")
			fakeKubernetesClientSet := NewFakeKubernetesClientSet(t)
			fakeKubernetesClientSet.On("GetJWT").Return("", fakeError).Once()

			authenticator := auth.NewKubernetesSecretProvider(
				"fooRole",
				auth.DefaultKubernetesAuthBackendName,
				fakeKubernetesClientSet,
			)

			secret, err := authenticator.GetSecret(fakeClient)

			assert.Equal(t, fakeError, stacktrace.RootCause(err))
			assert.Nil(t, secret)
			fakeClient.AssertExpectations(t)
			fakeKubernetesClientSet.AssertExpectations(t)
		},
	)

	t.Run(
		"when the kubernetes client returns empty JWT token, it returns nil secret and an error",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)

			fakeKubernetesClientSet := NewFakeKubernetesClientSet(t)
			fakeKubernetesClientSet.On("GetJWT").Return("", nil).Once()

			authenticator := auth.NewKubernetesSecretProvider(
				"fooRole",
				auth.DefaultKubernetesAuthBackendName,
				fakeKubernetesClientSet,
			)

			secret, err := authenticator.GetSecret(fakeClient)

			assert.Equal(t, "empty JWT service account JWT token", stacktrace.RootCause(err).Error())
			assert.Nil(t, secret)
			fakeClient.AssertExpectations(t)
			fakeKubernetesClientSet.AssertExpectations(t)
		},
	)
}
