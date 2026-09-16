// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package seclib

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestDecodePrivateKeyAccepted(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}
	pkcs1 := x509.MarshalPKCS1PrivateKey(key)

	cases := []struct {
		name string
		in   []byte
	}{
		{"pkcs8", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8})},
		{"pkcs1", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkcs1})},
	}
	for _, c := range cases {
		got, err := DecodePrivateKey(c.in)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", c.name, err)
			continue
		}
		if !got.Equal(key) {
			t.Errorf("%s: decoded key differs from the original", c.name)
		}
	}
}

// Input that is not PEM, or that carries a key which is not RSA, must produce
// an error rather than a partially decoded key.
func TestDecodePrivateKeyRejected(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("ecdsa.GenerateKey: %v", err)
	}
	ecPKCS8, err := x509.MarshalPKCS8PrivateKey(ecKey)
	if err != nil {
		t.Fatalf("MarshalPKCS8PrivateKey: %v", err)
	}

	cases := []struct {
		name string
		in   []byte
	}{
		{"nil input", nil},
		{"empty input", []byte{}},
		{"not pem", []byte("not a pem at all\n")},
		{"pem with unparsable body", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("garbage")})},
		{"ec pkcs8", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: ecPKCS8})},
	}
	for _, c := range cases {
		key, err := DecodePrivateKey(c.in)
		if err == nil {
			t.Errorf("%s: expected an error, got key %v", c.name, key)
		}
		if key != nil {
			t.Errorf("%s: expected a nil key when returning an error", c.name)
		}
	}
}
