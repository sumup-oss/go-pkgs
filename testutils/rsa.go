// Copyright 2019 SumUp Ltd.
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

package testutils

import (
	"crypto/rand"
	stdRsa "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func GenerateAndWritePrivateKey(t *testing.T, tmpDir, keyName string) (string, *stdRsa.PrivateKey) {
	privKey, err := stdRsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)

	keyPath := filepath.Join(tmpDir, keyName)

	pemBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privKey),
		},
	)

	err = ioutil.WriteFile(keyPath, pemBytes, 0600)
	require.Nil(t, err)

	return keyPath, privKey
}

func GenerateAndWritePublicKey(t *testing.T, tmpDir, keyName string, privKey *stdRsa.PrivateKey) string {
	keyPath := filepath.Join(tmpDir, keyName)

	pubkeyBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	require.Nil(t, err)

	pemBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubkeyBytes,
		},
	)

	err = ioutil.WriteFile(keyPath, pemBytes, 0644)
	require.Nil(t, err)
	return keyPath
}
