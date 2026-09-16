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

package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecryptAESGCM(t *testing.T) {
	cases := []struct {
		name string
		enc  func(string, string) (string, error)
		dec  func(string, string) (string, error)
	}{
		{"base64", EncryptAESBase64GCM, DescryptAESBase64GCM},
		{"base64url", EncryptAESBase64UrlGCM, DescryptAESBase64UrlGCM},
	}
	msgs := []string{"", "hello", "中文消息", strings.Repeat("x", 4096)}
	for _, c := range cases {
		for _, msg := range msgs {
			ct, err := c.enc("123", msg)
			if err != nil {
				t.Fatalf("%s: encrypt: %v", c.name, err)
			}
			got, err := c.dec("123", ct)
			if err != nil {
				t.Fatalf("%s: decrypt: %v", c.name, err)
			}
			if got != msg {
				t.Errorf("%s: round trip = %q, want %q", c.name, got, msg)
			}
		}
	}
}

// Two encryptions of the same message must not produce the same payload.
func TestEncryptAESGCMUsesFreshNonce(t *testing.T) {
	first, err := EncryptAESBase64GCM("123", "hello")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	second, err := EncryptAESBase64GCM("123", "hello")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if first == second {
		t.Error("two encryptions of the same message produced identical payloads")
	}
}

func TestDescryptAESGCMRejectsTamperedPayload(t *testing.T) {
	ct, err := EncryptAESBase64GCM("123", "hello world")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(ct)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	positions := []int{0, len(aesGCMHeader) - 1, len(aesGCMHeader) + 4, len(raw) - 1}
	for _, pos := range positions {
		modified := append([]byte(nil), raw...)
		modified[pos] ^= 0x01
		tampered := base64.StdEncoding.EncodeToString(modified)
		if _, err := DescryptAESBase64GCM("123", tampered); err == nil {
			t.Errorf("modifying byte %d was not reported", pos)
		}
	}
}

func TestDescryptAESGCMRejectsWrongKey(t *testing.T) {
	ct, err := EncryptAESBase64GCM("123", "hello")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if _, err := DescryptAESBase64GCM("456", ct); err == nil {
		t.Error("decrypting with the wrong key was not reported")
	}
}

// The Auto functions read both the authenticated payloads and the legacy ones.
func TestDescryptAESAutoReadsBothFormats(t *testing.T) {
	const key = "123"
	const msg = "migrated payload"

	legacyStd, err := EncryptAESBase64(key, msg)
	if err != nil {
		t.Fatalf("EncryptAESBase64: %v", err)
	}
	legacyURL, err := EncryptAESBase64Url(key, msg)
	if err != nil {
		t.Fatalf("EncryptAESBase64Url: %v", err)
	}
	gcmStd, err := EncryptAESBase64GCM(key, msg)
	if err != nil {
		t.Fatalf("EncryptAESBase64GCM: %v", err)
	}
	gcmURL, err := EncryptAESBase64UrlGCM(key, msg)
	if err != nil {
		t.Fatalf("EncryptAESBase64UrlGCM: %v", err)
	}

	cases := []struct {
		name string
		dec  func(string, string) (string, error)
		in   string
	}{
		{"auto<-legacy", DescryptAESBase64Auto, legacyStd},
		{"auto<-gcm", DescryptAESBase64Auto, gcmStd},
		{"autoUrl<-legacy", DescryptAESBase64UrlAuto, legacyURL},
		{"autoUrl<-gcm", DescryptAESBase64UrlAuto, gcmURL},
	}
	for _, c := range cases {
		got, err := c.dec(key, c.in)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if got != msg {
			t.Errorf("%s: got %q, want %q", c.name, got, msg)
		}
	}
}

// The strict GCM reader must not accept a legacy payload, and the legacy reader
// keeps working on payloads written before this change.
func TestFormatReadersAreDistinct(t *testing.T) {
	legacy, err := EncryptAESBase64("123", "hello")
	if err != nil {
		t.Fatalf("EncryptAESBase64: %v", err)
	}
	if _, err := DescryptAESBase64GCM("123", legacy); err == nil {
		t.Error("the GCM reader accepted a legacy payload")
	}

	// Payload written by the previous revision, kept as a fixture so the
	// legacy reader stays compatible with already stored values.
	if _, err := DescryptAESBase64("123", "zMuWP5HwnC+zqNayjZGSouZCHA=="); err != nil {
		t.Errorf("legacy fixture no longer decrypts: %v", err)
	}
}
