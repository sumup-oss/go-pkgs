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
	"github.com/stretchr/testify/mock"

	"github.com/sumup-oss/go-pkgs/executor/vault/auth"
)

type FakeSecretProvider struct {
	mock.Mock
}

func NewFakeSecretProvider(t *testing.T) *FakeSecretProvider {
	fake := &FakeSecretProvider{}
	fake.Test(t)

	return fake
}

func (f *FakeSecretProvider) GetSecret(client auth.Client) (*vaultapi.Secret, error) {
	args := f.Called(client)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*vaultapi.Secret), args.Error(1)
}

func TestAuthenticator_Authenticate(t *testing.T) {
	t.Run(
		"when there is no current token, "+
			"it must renew the token and set the token to the provided vault client",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("SetToken", "tokenID").Once()

			secret := &vaultapi.Secret{
				Data: map[string]interface{}{
					"id": "tokenID",
				},
			}
			fakeSecretProvider := NewFakeSecretProvider(t)
			fakeSecretProvider.On("GetSecret", fakeClient).Return(secret, nil)

			authenticator := auth.NewAuthenticator(fakeSecretProvider)

			err := authenticator.Authenticate(fakeClient)

			assert.Nil(t, err)
			fakeClient.AssertExpectations(t)
			fakeSecretProvider.AssertExpectations(t)
		},
	)

	t.Run(
		"when there is current unexpired token, "+
			"it must set the current token to the provided vault client",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("SetToken", "tokenID").Once()

			secret := &vaultapi.Secret{
				Data: map[string]interface{}{
					"id":  "tokenID",
					"ttl": "10s",
				},
			}
			fakeSecretProvider := NewFakeSecretProvider(t)
			fakeSecretProvider.On("GetSecret", fakeClient).Return(secret, nil).Once()

			authenticator := auth.NewAuthenticator(fakeSecretProvider)

			// first call uses the secret
			err := authenticator.Authenticate(fakeClient)
			assert.Nil(t, err)

			// second call reuses the old secret, meaning that it does nothing
			err = authenticator.Authenticate(fakeClient)
			assert.Nil(t, err)

			fakeClient.AssertExpectations(t)
			fakeSecretProvider.AssertExpectations(t)
		},
	)

	t.Run(
		"when the current token is expired, "+
			"it must renew the token and set the new token to the provided vault client",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("SetToken", "expiredTokenID").Once()
			fakeClient.On("SetToken", "newTokenID").Once()

			expiredSecret := &vaultapi.Secret{
				Data: map[string]interface{}{
					"id":  "expiredTokenID",
					"ttl": "4s",
				},
			}
			newSecret := &vaultapi.Secret{
				Data: map[string]interface{}{
					"id":  "newTokenID",
					"ttl": "10s",
				},
			}
			fakeSecretProvider := NewFakeSecretProvider(t)
			fakeSecretProvider.On("GetSecret", fakeClient).Return(expiredSecret, nil).Once()
			fakeSecretProvider.On("GetSecret", fakeClient).Return(newSecret, nil).Once()

			authenticator := auth.NewAuthenticator(fakeSecretProvider)

			// first call uses the expiredSecret
			err := authenticator.Authenticate(fakeClient)
			assert.Nil(t, err)

			// second call creates newSecret
			err = authenticator.Authenticate(fakeClient)
			assert.Nil(t, err)

			// third call reuses the newSecret, meaning that it does nothing
			err = authenticator.Authenticate(fakeClient)
			assert.Nil(t, err)

			fakeClient.AssertExpectations(t)
			fakeSecretProvider.AssertExpectations(t)
		},
	)

	t.Run(
		"when fetching a secret from the secret provider fails, "+
			"it must not set any token to the vault client and return an error",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)

			fakeError := errors.New("fakeError")
			fakeSecretProvider := NewFakeSecretProvider(t)
			fakeSecretProvider.On("GetSecret", fakeClient).Return(nil, fakeError).Once()

			authenticator := auth.NewAuthenticator(fakeSecretProvider)

			err := authenticator.Authenticate(fakeClient)

			assert.Equal(t, fakeError, stacktrace.RootCause(err))
			fakeClient.AssertExpectations(t)
			fakeSecretProvider.AssertExpectations(t)
		},
	)
}
