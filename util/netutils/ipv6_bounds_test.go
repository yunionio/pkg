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

import "testing"

// An address with more groups than a 128 bit address can hold has no valid
// reading and must be reported as an error.
func TestNewIPV6AddrRejectsTooManyGroups(t *testing.T) {
	inputs := []string{
		"1:2:3:4:5:6:7:8:9",
		"1:2:3:4:5:6:7:8:9:a",
		"::::::::1",
		":::::::::1",
		"1:2:3:4:5:6:7:1.2.3.4",
		"1:2:3:4:5:6:7:8:1.2.3.4",
	}
	for _, in := range inputs {
		addr, err := NewIPV6Addr(in)
		if err == nil {
			t.Errorf("NewIPV6Addr(%q) returned no error, got %v", in, addr)
		}
	}
}

// ParsePrefix6 goes through the same parser, so it must reject them too.
func TestParsePrefix6RejectsTooManyGroups(t *testing.T) {
	inputs := []string{
		"1:2:3:4:5:6:7:8:9/64",
		"1:2:3:4:5:6:7:8:9:a/64",
	}
	for _, in := range inputs {
		if _, _, err := ParsePrefix6(in); err == nil {
			t.Errorf("ParsePrefix6(%q) returned no error", in)
		}
	}
}

// Addresses that were valid before must still parse to the same value.
func TestNewIPV6AddrStillAcceptsValidInput(t *testing.T) {
	cases := []struct {
		in   string
		want IPV6Addr
	}{
		{"::", IPV6Addr{}},
		{"::1", IPV6Addr{0, 0, 0, 0, 0, 0, 0, 1}},
		{"1:2:3:4:5:6:7:8", IPV6Addr{1, 2, 3, 4, 5, 6, 7, 8}},
		{"fe80::1", IPV6Addr{0xfe80, 0, 0, 0, 0, 0, 0, 1}},
		{"::192.168.0.1", IPV6Addr{0, 0, 0, 0, 0, 0, 0xc0a8, 1}},
		{"2001:db8:3333:4444:5555:6666:1.2.3.4",
			IPV6Addr{0x2001, 0xdb8, 0x3333, 0x4444, 0x5555, 0x6666, 0x0102, 0x0304}},
	}
	for _, c := range cases {
		got, err := NewIPV6Addr(c.in)
		if err != nil {
			t.Errorf("NewIPV6Addr(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("NewIPV6Addr(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
