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

// A pair that cannot be decoded must not stop the pairs around it from being
// masked, and must not cause the raw body to be used instead of the masked
// one. url.ParseQuery does return the pairs it did decode alongside its error,
// so taking only its return value would either drop them or leak them.
func TestRedactFormBodyWithUndecodablePair(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"bad escape first", "bad=%ZZ&username=admin&password=hunter2"},
		{"bad escape last", "username=admin&password=hunter2&bad=%ZZ"},
		{"bad escape in the middle", "username=admin&bad=%ZZ&password=hunter2"},
		{"percent encoded name", "username=admin&%70assword=hunter2"},
		{"pair without a value", "username=admin&password&flag"},
	}
	for _, c := range cases {
		msg := errorBody(t, "application/x-www-form-urlencoded", c.body)
		if strings.Contains(msg, "hunter2") {
			t.Errorf("%s: password survived into the message: %s", c.name, msg)
		}
		if !strings.Contains(msg, "admin") {
			t.Errorf("%s: the non-sensitive value was lost: %s", c.name, msg)
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
			name:    "attribute, double quoted",
			body:    `<Login Username="admin" SecretKey="shh123" />`,
			secrets: []string{"shh123"},
		},
		{
			name:    "attribute, single quoted",
			body:    `<Login Username='admin' SecretKey='shh123' />`,
			secrets: []string{"shh123"},
		},
		{
			name:    "attribute, unquoted is not an attribute value",
			body:    `<Login Username="admin" SecretKey="shh123">x</Login>`,
			secrets: []string{"shh123"},
		},
		{
			name:    "CDATA section",
			body:    `<Login><Username>admin</Username><Password><![CDATA[hunter2]]></Password></Login>`,
			secrets: []string{"hunter2"},
		},
		{
			name:    "CDATA across lines",
			body:    "<Login><Username>admin</Username><Password><![CDATA[line1\nhunter2\n]]></Password></Login>",
			secrets: []string{"hunter2"},
		},
		{
			name:    "empty element then a real one",
			body:    `<Login><Username>admin</Username><Password></Password><Token>abc123</Token></Login>`,
			secrets: []string{"abc123"},
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

// A body that cannot be walked field by field is left out, whatever its size.
//
// A multipart body names its fields in a header line, so masking the names
// would leave the values readable — the body has to go rather than be masked.
func TestUninterpretableBodyIsOmitted(t *testing.T) {
	cases := []struct {
		name     string
		contType string
		body     string
		secret   string
	}{
		{
			name:     "multipart",
			contType: "multipart/form-data; boundary=----b",
			body:     "------b\r\nContent-Disposition: form-data; name=\"password\"\r\n\r\nhunter2\r\n------b--\r\n",
			secret:   "hunter2",
		},
		{
			name:     "small unknown type",
			contType: "application/octet-stream",
			body:     "small",
			secret:   "small",
		},
		{
			name:     "large unknown type",
			contType: "application/octet-stream",
			body:     strings.Repeat("Z", 4096),
			secret:   "ZZZZ",
		},
		{
			name:     "no content type",
			contType: "",
			body:     `{"username":"admin","password":"hunter2"}`,
			secret:   "hunter2",
		},
	}
	for _, c := range cases {
		msg := errorBody(t, c.contType, c.body)
		if strings.Contains(msg, c.secret) {
			t.Errorf("%s: %q was reproduced in the message: %s", c.name, c.secret, msg)
		}
	}
}

// A body larger than the bound is left out rather than cut down, because a cut
// keeps whichever part happened to fall inside it.
func TestOversizedBodyIsOmitted(t *testing.T) {
	for _, contType := range []string{"application/json", "application/xml", "application/x-www-form-urlencoded"} {
		body := strings.Repeat("a", maxRedactedBodyBytes+1)
		msg := errorBody(t, contType, body)
		if strings.Contains(msg, "aaaa") {
			t.Errorf("%s: an oversized body was included: %s", contType, msg)
		}
	}
	// Just inside the bound is still included.
	msg := errorBody(t, "application/x-www-form-urlencoded", "password=x&"+strings.Repeat("a", 16))
	if strings.Contains(msg, "password=x") {
		t.Errorf("a body inside the bound was not masked: %s", msg)
	}
}

// A JSON body that does not parse still has its field names in the text, so
// the values can still be masked.
func TestMalformedJSONBodyIsStillMasked(t *testing.T) {
	msg := errorBody(t, "application/json", `{"username":"admin","password":"hunter2",`)
	if strings.Contains(msg, "hunter2") {
		t.Errorf("password survived into the message: %s", msg)
	}
	if !strings.Contains(msg, "admin") {
		t.Errorf("the non-sensitive value was lost: %s", msg)
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
