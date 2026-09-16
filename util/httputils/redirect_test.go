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
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// redirectingServer serves /start by redirecting to the same address under the
// given host name, and /end by recording the headers it was reached with.
//
// The same listener is reachable as both 127.0.0.1 and localhost, which lets a
// test redirect to a genuinely different host name without needing two
// servers.
func redirectingServer(host string, reached *http.Header) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			u, err := url.Parse(srv.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			u.Host = host + ":" + u.Port()
			http.Redirect(w, r, u.String()+"/end", http.StatusFound)
			return
		}
		*reached = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
	}))
	return srv
}

// A redirect to a different host name must not carry the session headers.
func TestRedirectToOtherHostDropsSessionHeaders(t *testing.T) {
	var reached http.Header
	srv := redirectingServer("localhost", &reached)
	defer srv.Close()

	client := GetTimeoutClient(5 * time.Second)
	req, err := http.NewRequest("GET", srv.URL+"/start", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Auth-Token", "tokensecret")
	req.Header.Set("X-Subject-Token", "subjectsecret")
	req.Header.Set("X-Trace", "tracevalue")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()

	if got := reached.Get("X-Auth-Token"); got != "" {
		t.Errorf("X-Auth-Token reached the other host: %q", got)
	}
	if got := reached.Get("X-Subject-Token"); got != "" {
		t.Errorf("X-Subject-Token reached the other host: %q", got)
	}
	// Everything else still travels.
	if got := reached.Get("X-Trace"); got != "tracevalue" {
		t.Errorf("X-Trace = %q, want it to survive the redirect", got)
	}
}

// A redirect to the same host name keeps them.
func TestRedirectWithinHostKeepsSessionHeaders(t *testing.T) {
	var reached http.Header
	srv := redirectingServer("127.0.0.1", &reached)
	defer srv.Close()

	client := GetTimeoutClient(5 * time.Second)
	req, err := http.NewRequest("GET", srv.URL+"/start", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Auth-Token", "tokensecret")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()

	if got := reached.Get("X-Auth-Token"); got != "tokensecret" {
		t.Errorf("X-Auth-Token = %q, want it to survive a same host redirect", got)
	}
}

func TestIsSameOrSubdomain(t *testing.T) {
	cases := []struct {
		initial string
		dest    string
		want    bool
	}{
		{"https://api.example.com", "https://api.example.com", true},
		{"https://api.example.com", "https://api.example.com:8443", true},
		{"https://example.com", "https://api.example.com", true},
		{"https://API.example.com", "https://api.example.com", true},
		// A redirect up to the parent domain is not covered.
		{"https://api.example.com", "https://example.com", false},
		{"https://api.example.com", "https://elsewhere.example.com", false},
		{"https://api.example.com", "https://evil.com", false},
		// A suffix that is not a label boundary.
		{"https://api.example.com", "https://notapi.example.com", false},
		{"https://example.com", "https://example.com.evil.com", false},
	}
	for _, c := range cases {
		i, err := url.Parse(c.initial)
		if err != nil {
			t.Fatalf("parse %q: %v", c.initial, err)
		}
		d, err := url.Parse(c.dest)
		if err != nil {
			t.Fatalf("parse %q: %v", c.dest, err)
		}
		if got := isSameOrSubdomain(i, d); got != c.want {
			t.Errorf("isSameOrSubdomain(%q, %q) = %v, want %v", c.initial, c.dest, got, c.want)
		}
	}
}

// The usual limit on how many redirects are followed still applies.
func TestRedirectLoopIsStopped(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL, http.StatusFound)
	}))
	defer srv.Close()

	client := GetTimeoutClient(5 * time.Second)
	resp, err := client.Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected the redirect loop to be stopped")
	}
}
