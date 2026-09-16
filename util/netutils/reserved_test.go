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

package netutils

import (
	"net/http/httptest"
	"testing"
)

func mustIPv4(t *testing.T, s string) IPV4Addr {
	t.Helper()
	addr, err := NewIPV4Addr(s)
	if err != nil {
		t.Fatalf("NewIPV4Addr(%q): %v", s, err)
	}
	return addr
}

func TestIsReserved(t *testing.T) {
	reserved := []string{
		"0.0.0.0", "0.1.0.1", "0.255.255.255",
		"192.0.2.1", "198.51.100.1", "203.0.113.1",
		"240.0.0.1", "255.255.255.255",
	}
	for _, s := range reserved {
		if !IsReserved(mustIPv4(t, s)) {
			t.Errorf("IsReserved(%s) = false, want true", s)
		}
	}
	notReserved := []string{
		"1.1.1.1", "8.8.8.8", "114.113.226.53",
		"10.0.0.1", "192.168.1.1", "127.0.0.1", "169.254.1.1", "224.0.0.1",
	}
	for _, s := range notReserved {
		if IsReserved(mustIPv4(t, s)) {
			t.Errorf("IsReserved(%s) = true, want false", s)
		}
	}
}

// A reserved address cannot be one that a guest reaches directly.
func TestIsExitAddressExcludesReserved(t *testing.T) {
	notExit := []string{
		"0.0.0.0", "0.1.0.1",
		"192.0.2.1", "198.51.100.1", "203.0.113.1",
		"240.0.0.1", "255.255.255.255",
		// already excluded before this change
		"10.0.0.1", "172.31.32.1", "192.168.222.177",
		"127.0.0.1", "169.254.1.1", "224.0.0.1",
	}
	for _, s := range notExit {
		if IsExitAddress(mustIPv4(t, s)) {
			t.Errorf("IsExitAddress(%s) = true, want false", s)
		}
	}
	exit := []string{"1.1.1.1", "8.8.8.8", "114.113.226.53", "100.63.0.1", "198.20.0.1"}
	for _, s := range exit {
		if !IsExitAddress(mustIPv4(t, s)) {
			t.Errorf("IsExitAddress(%s) = false, want true", s)
		}
	}
}

// IsPrivate keeps the meaning it had: the reserved ranges are not folded
// into it, because allocation and DHCP callers classify on it.
func TestIsPrivateUnchangedByReservedRanges(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"100.64.0.1", true},
		// reserved, but not private
		{"0.1.0.1", false},
		{"192.0.2.1", false},
		{"240.0.0.1", false},
		{"1.1.1.1", false},
	}
	for _, c := range cases {
		if got := IsPrivate(mustIPv4(t, c.addr)); got != c.want {
			t.Errorf("IsPrivate(%s) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestGetHttpRequestIp(t *testing.T) {
	cases := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{"forwarded", "203.0.113.7", "", "10.0.0.1:1234", "203.0.113.7"},
		{"forwarded with spaces", "  203.0.113.7  ", "", "10.0.0.1:1234", "203.0.113.7"},
		{"forwarded list", "203.0.113.7, 10.0.0.1", "", "10.0.0.1:1234", "203.0.113.7"},
		{"forwarded not an address", "not-an-ip", "", "10.0.0.1:1234", "10.0.0.1"},
		{"forwarded not an address, real ip set", "not-an-ip", "198.51.100.9", "10.0.0.1:1234", "198.51.100.9"},
		{"real ip", "", "198.51.100.9", "10.0.0.1:1234", "198.51.100.9"},
		{"real ip not an address", "", "garbage", "10.0.0.1:1234", "10.0.0.1"},
		{"remote addr ipv4", "", "", "10.0.0.1:1234", "10.0.0.1"},
		{"remote addr ipv6", "", "", "[2001:db8::1]:1234", "2001:db8::1"},
		{"remote addr loopback ipv6", "", "", "[::1]:1234", "::1"},
		{"remote addr without port", "", "", "10.0.0.1", "10.0.0.1"},
	}
	for _, c := range cases {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = c.remoteAddr
		if len(c.xff) > 0 {
			req.Header.Set("X-Forwarded-For", c.xff)
		}
		if len(c.xri) > 0 {
			req.Header.Set("X-Real-Ip", c.xri)
		}
		if got := GetHttpRequestIp(req); got != c.want {
			t.Errorf("%s: GetHttpRequestIp() = %q, want %q", c.name, got, c.want)
		}
	}
}
