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

func TestRedactURL(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		mustNot  []string // substrings that must not survive
		mustKeep []string // substrings that must survive
	}{
		{
			name:     "presigned url",
			in:       "https://bucket.example.com/obj?X-Amz-Credential=AKIA/20260101&X-Amz-Signature=deadbeef&X-Amz-Date=20260101T000000Z",
			mustNot:  []string{"AKIA", "deadbeef"},
			mustKeep: []string{"X-Amz-Credential", "X-Amz-Signature", "bucket.example.com", "X-Amz-Date"},
		},
		{
			name:     "plain token",
			in:       "https://api.example.com/v1/things?token=abc123&limit=10",
			mustNot:  []string{"abc123"},
			mustKeep: []string{"token", "limit=10", "api.example.com"},
		},
		{
			name:     "access token",
			in:       "https://api.example.com/v1/things?access_token=abc123",
			mustNot:  []string{"abc123"},
			mustKeep: []string{"access_token"},
		},
		{
			name:     "oauth code",
			in:       "https://idp.example.com/callback?code=SplxlOBeZQQYbYS6WxSbIA&state=x",
			mustNot:  []string{"SplxlOBeZQQYbYS6WxSbIA"},
			mustKeep: []string{"code", "state=x"},
		},
		{
			name:     "password",
			in:       "https://api.example.com/login?password=hunter2",
			mustNot:  []string{"hunter2"},
			mustKeep: []string{"password"},
		},
		{
			name:     "api key",
			in:       "https://api.example.com/v1?apikey=secretvalue",
			mustNot:  []string{"secretvalue"},
			mustKeep: []string{"apikey"},
		},
		{
			name:     "secret",
			in:       "https://api.example.com/v1?client_secret=shh123",
			mustNot:  []string{"shh123"},
			mustKeep: []string{"client_secret"},
		},
		{
			// An ordinary URL is returned exactly as it was, so nothing
			// that was already there is reformatted.
			name:     "nothing to mask",
			in:       "https://api.example.com/v1/things?limit=10&marker=a%2Fb",
			mustKeep: []string{"limit=10", "marker=a%2Fb"},
		},
		{
			name:     "no query",
			in:       "https://api.example.com/v1/things",
			mustKeep: []string{"https://api.example.com/v1/things"},
		},
	}
	for _, c := range cases {
		got := redactURL(c.in)
		for _, s := range c.mustNot {
			if strings.Contains(got, s) {
				t.Errorf("%s: %q still present in %q", c.name, s, got)
			}
		}
		for _, s := range c.mustKeep {
			if !strings.Contains(got, s) {
				t.Errorf("%s: %q missing from %q", c.name, s, got)
			}
		}
	}
}

func TestRedactURLLeavesCleanURLByteForByte(t *testing.T) {
	const in = "https://api.example.com/v1/things?limit=10&marker=a%2Fb"
	if got := redactURL(in); got != in {
		t.Errorf("redactURL(%q) = %q, want it unchanged", in, got)
	}
}

func TestNewJsonClientErrorFromRequest2Redacts(t *testing.T) {
	hdrs := http.Header{}
	hdrs.Set("Content-Type", "application/json")
	hdrs.Set("Authorization", "Bearer topsecret")
	hdrs.Set("X-Auth-Token", "tokensecret")
	hdrs.Set("X-Subject-Token", "subjectsecret")
	hdrs.Set("Cookie", "session=cookiesecret")
	hdrs.Set("X-Trace", "tracevalue")

	ce := newJsonClientErrorFromRequest2("GET",
		"https://api.example.com/v1/things?token=querysecret&limit=10", hdrs, "")

	body := ce.Error()
	for _, secret := range []string{
		"topsecret", "tokensecret", "subjectsecret", "cookiesecret", "querysecret",
	} {
		if strings.Contains(body, secret) {
			t.Errorf("%q survived into the error message: %s", secret, body)
		}
	}
	for _, kept := range []string{"limit=10", "tracevalue", "api.example.com"} {
		if !strings.Contains(body, kept) {
			t.Errorf("%q missing from the error message: %s", kept, body)
		}
	}
}
