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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumup-oss/go-pkgs/executor/vault/auth"
)

func TestTokenAuthenticator_Authenticate(t *testing.T) {
	t.Run(
		"it uses the fixed token to authenticate the provided vault client and returns no error",
		func(t *testing.T) {
			fakeClient := NewFakeVaultClient(t)
			fakeClient.On("SetToken", "fooToken").Once()
			authenticator := auth.NewTokenAuthenticator("fooToken")

			err := authenticator.Authenticate(fakeClient)

			assert.Nil(t, err)
			fakeClient.AssertExpectations(t)
		},
	)
}
