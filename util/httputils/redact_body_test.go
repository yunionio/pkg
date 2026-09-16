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

package httputils

import (
	"net/http"
	"strings"
	"testing"
)

func errorBody(t *testing.T, contType, body string) string {
	t.Helper()
	hdrs := http.Header{}
	hdrs.Set("Content-Type", contType)
	ce := newJsonClientErrorFromRequest2("POST", "https://api.example.com/v1/things", hdrs, body)
	return ce.Error()
}

func TestRedactJSONBody(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		secrets []string
		keep    []string
	}{
		{
			name:    "password",
			body:    `{"username":"admin","password":"hunter2"}`,
			secrets: []string{"hunter2"},
			keep:    []string{"admin", "username", "password"},
		},
		{
			name:    "new_password is not the same key",
			body:    `{"username":"admin","new_password":"hunter2"}`,
			secrets: []string{"hunter2"},
			keep:    []string{"admin", "new_password"},
		},
		{
			name:    "secret key",
			body:    `{"access_key_id":"AKIA","secret_key":"shh123"}`,
			secrets: []string{"shh123"},
			keep:    []string{"AKIA", "access_key_id"},
		},
		{
			name:    "token",
			body:    `{"token":"abc123","limit":10}`,
			secrets: []string{"abc123"},
			keep:    []string{"limit", "10"},
		},
		{
			name:    "nested",
			body:    `{"user":{"name":"admin","password":"hunter2"},"note":"hi"}`,
			secrets: []string{"hunter2"},
			keep:    []string{"admin", "note", "hi"},
		},
		{
			name:    "inside an array",
			body:    `{"items":[{"id":1,"token":"abc123"},{"id":2}]}`,
			secrets: []string{"abc123"},
			keep:    []string{"items"},
		},
	}
	for _, c := range cases {
		msg := errorBody(t, "application/json", c.body)
		for _, s := range c.secrets {
			if strings.Contains(msg, s) {
				t.Errorf("%s: %q survived into the message: %s", c.name, s, msg)
			}
		}
		for _, s := range c.keep {
			if !strings.Contains(msg, s) {
				t.Errorf("%s: %q missing from the message: %s", c.name, s, msg)
			}
		}
	}
}

// A form encoded body is walked by field name, which the previous behaviour
// did not do at all.
func TestRedactFormBody(t *testing.T) {
	msg := errorBody(t, "application/x-www-form-urlencoded", "username=admin&password=hunter2&limit=10")
	if strings.Contains(msg, "hunter2") {
		t.Errorf("form password survived into the message: %s", msg)
	}
	for _, s := range []string{"username", "admin", "limit", "10"} {
		if !strings.Contains(msg, s) {
			t.Errorf("form %q missing from the message: %s", s, msg)
		}
	}
}

func TestRedactXMLBody(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		secrets []string
	}{
		{
			name:    "element",
			body:    `<Login><Username>admin</Username><Password>hunter2</Password></Login>`,
			secrets: []string{"hunter2"},
		},
		{
			name:    "attribute",
			body:    `<Login Username="admin" SecretKey="shh123" />`,
			secrets: []string{"shh123"},
		},
	}
	for _, c := range cases {
		msg := errorBody(t, "application/xml", c.body)
		for _, s := range c.secrets {
			if strings.Contains(msg, s) {
				t.Errorf("%s: %q survived into the message: %s", c.name, s, msg)
			}
		}
		if !strings.Contains(msg, "admin") {
			t.Errorf("%s: the non-sensitive value was lost: %s", c.name, msg)
		}
	}
}

// A body that cannot be walked field by field is only included as a bounded
// excerpt, whatever its content type.
func TestUninterpretableBodyIsTruncated(t *testing.T) {
	body := strings.Repeat("Z", 4096)
	msg := errorBody(t, "application/octet-stream", body)
	if strings.Contains(msg, body) {
		t.Error("the whole body was reproduced in the message")
	}
	if !strings.Contains(msg, "...") {
		t.Error("the excerpt is not marked as truncated")
	}
}

// A small body of a type this package does not understand is left out.
func TestUnknownContentTypeSmallBodyIsOmitted(t *testing.T) {
	msg := errorBody(t, "application/octet-stream", "small")
	if strings.Contains(msg, "small") {
		t.Errorf("an uninterpretable body was included: %s", msg)
	}
}

// Rendering the error twice gives the same message and does not change the
// error it carries.
func TestJSONClientErrorErrorIsIdempotent(t *testing.T) {
	hdrs := http.Header{}
	hdrs.Set("Content-Type", "application/json")
	ce := newJsonClientErrorFromRequest2("POST", "https://api.example.com/v1",
		hdrs, `{"password":"hunter2","user":"admin"}`)

	first := ce.Error()
	second := ce.Error()
	if first != second {
		t.Errorf("the message changed between calls:\n%s\n%s", first, second)
	}
}
