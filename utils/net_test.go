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

func TestGetAddrPort(t *testing.T) {
	cases := []struct {
		in       string
		wantHost string
		wantPort int
	}{
		{"1.2.3.4:80", "1.2.3.4", 80},
		{"192.168.1.1:65535", "192.168.1.1", 65535},
		{"[::1]:8080", "::1", 8080},
		{"[2001:db8::1]:443", "2001:db8::1", 443},

		// Values that cannot be split must not take the caller down.
		{"", "", 0},
		{"1.2.3.4", "1.2.3.4", 0},
		{"x", "x", 0},
		{"1.2.3.4:", "1.2.3.4", 0},
		{"1.2.3.4:abc", "1.2.3.4", 0},
		{"::1", "::1", 0},
	}
	for _, c := range cases {
		host, port := GetAddrPort(c.in)
		if host != c.wantHost || port != c.wantPort {
			t.Errorf("GetAddrPort(%q) = (%q, %d), want (%q, %d)",
				c.in, host, port, c.wantHost, c.wantPort)
		}
	}
}
