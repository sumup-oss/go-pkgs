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
	"sync"
	"time"

	"github.com/palantir/stacktrace"
)

// Authenticator authenticates a vault client using a given SecretProvider strategy.
//
// The Authenticator renews the currently issued vault auth token on demand at least 5 seconds
// before the last issued token expires.
type Authenticator struct {
	secretProvider SecretProvider

	// protects vaultToken
	mu         sync.RWMutex
	vaultToken *TokenInfo
}

func NewAuthenticator(secretProvider SecretProvider) *Authenticator {
	return &Authenticator{
		secretProvider: secretProvider,
	}
}

// Authenticate authenticates the vault client.
func (a *Authenticator) Authenticate(client Client) error {
	currentToken := a.getToken()
	if currentToken != nil &&
		time.Now().UTC().Before(currentToken.TokenExpireTimeUTC.Add(-5*time.Second)) {
		return nil
	}

	return a.renewToken(client)
}

func (a *Authenticator) renewToken(client Client) error {
	tokenIssueTimeUTC := time.Now().UTC()

	secret, err := a.secretProvider.GetSecret(client)
	if err != nil {
		return stacktrace.Propagate(err, "failed to login in Vault")
	}

	tokenID, err := secret.TokenID()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get Vault Token id")
	}

	tokenAccessor, err := secret.TokenAccessor()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get Vault Token accessor")
	}

	tokenTTL, err := secret.TokenTTL()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get Vault Token TTL")
	}

	tokenIsRenewable, err := secret.TokenIsRenewable()
	if err != nil {
		return stacktrace.Propagate(err, "failed to get Vault Token renewable")
	}

	tokenInfo := &TokenInfo{
		TokenID:            tokenID,
		TokenAccessor:      tokenAccessor,
		TokenDuration:      tokenTTL,
		TokenExpireTimeUTC: tokenIssueTimeUTC.Add(tokenTTL),
		TokenRenewable:     tokenIsRenewable,
	}

	a.setToken(tokenInfo)
	client.SetToken(tokenID)

	return nil
}

func (a *Authenticator) setToken(token *TokenInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.vaultToken = token
}

func (a *Authenticator) getToken() *TokenInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.vaultToken
}
