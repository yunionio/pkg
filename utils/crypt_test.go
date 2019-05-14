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

import "testing"

func TestEncryptAESBase64(t *testing.T) {
	key := "123"
	lMsg := "This is the secret!"

	cypher, err := EncryptAESBase64(key, lMsg)
	if err != nil {
		t.Errorf("EncryptAESBase64 %s", err)
	} else {
		t.Logf("EncryptAESBase64 %s", cypher)
	}

	msg, err := DescryptAESBase64(key, cypher)
	if err != nil {
		t.Errorf("DescryptAESBase64 %s", err)
	} else {
		t.Logf("DescryptAESBase64 %s", msg)
	}

	if msg != lMsg {
		t.Errorf("%s != %s", msg, lMsg)
	}
}

func TestDescryptAESBase64(t *testing.T) {
	msg, err := DescryptAESBase64("123", "zMuWP5HwnC+zqNayjZGSouZCHA==")
	if err != nil {
		t.Errorf("DescryptAESBase64 %s", err)
	} else {
		t.Logf("DescryptAESBase64 %s", msg)
	}
}

func TestEncryptAESBase64Url(t *testing.T) {
	key := "mysecret"
	lMsg := "This is a long long cypher text msg!!!"

	cypher, err := EncryptAESBase64Url(key, lMsg)
	if err != nil {
		t.Errorf("EncryptAESBase64Url %s", err)
	} else {
		t.Logf("EncryptAESBase64Url: %s", cypher)
	}

	msg, err := DescryptAESBase64Url(key, cypher)
	if err != nil {
		t.Errorf("DescryptAESBase64Url %s", err)
	} else {
		t.Logf("DescryptAESBase64Url %s", msg)
	}

	if msg != lMsg {
		t.Errorf("%s != %s", msg, lMsg)
	}
}
