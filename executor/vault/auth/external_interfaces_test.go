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

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/mock"
)

type FakeVaultClient struct {
	mock.Mock
}

func NewFakeVaultClient(t *testing.T) *FakeVaultClient {
	fakeClient := &FakeVaultClient{}
	fakeClient.Test(t)

	return fakeClient
}

func (c *FakeVaultClient) SetToken(token string) {
	c.Called(token)
}

func (c *FakeVaultClient) Read(path string) (*vaultapi.Secret, error) {
	args := c.Called(path)
	return args.Get(0).(*vaultapi.Secret), args.Error(1)
}

func (c *FakeVaultClient) RawWrite(path string, data map[string]interface{}) (*vaultapi.Secret, error) {
	args := c.Called(path, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*vaultapi.Secret), args.Error(1)
}

type FakeKubernetesClientSet struct {
	mock.Mock
}

func NewFakeKubernetesClientSet(t *testing.T) *FakeKubernetesClientSet {
	fake := &FakeKubernetesClientSet{}
	fake.Test(t)

	return fake
}

func (f *FakeKubernetesClientSet) GetJWT() (string, error) {
	args := f.Called()
	return args.String(0), args.Error(1)
}
