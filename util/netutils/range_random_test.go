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

func mustAddr(t *testing.T, s string) IPV4Addr {
	t.Helper()
	addr, err := NewIPV4Addr(s)
	if err != nil {
		t.Fatalf("NewIPV4Addr(%q): %v", s, err)
	}
	return addr
}

// A range covering exactly one address has nothing to choose from.
func TestIPV4AddrRangeRandomSingleAddress(t *testing.T) {
	for _, s := range []string{"10.0.0.1", "0.0.0.0", "255.255.255.254"} {
		addr := mustAddr(t, s)
		r := NewIPV4AddrRange(addr, addr)
		for i := 0; i < 10; i++ {
			if got := r.Random(); got != addr {
				t.Fatalf("Random() over %s = %s, want %s", r, got, addr)
			}
		}
	}
}

// The zero value has start == end and must behave the same way.
func TestIPV4AddrRangeRandomZeroValue(t *testing.T) {
	var r IPV4AddrRange
	if got := r.Random(); got != r.start {
		t.Errorf("Random() over the zero range = %s, want %s", got, r.start)
	}
}

// Every draw must stay inside the range, and both ends must eventually appear.
func TestIPV4AddrRangeRandomStaysInRange(t *testing.T) {
	start := mustAddr(t, "192.168.20.100")
	end := mustAddr(t, "192.168.20.150")
	r := NewIPV4AddrRange(start, end)

	seen := map[IPV4Addr]bool{}
	for i := 0; i < 5000; i++ {
		got := r.Random()
		if got < start || got > end {
			t.Fatalf("Random() = %s outside %s", got, r)
		}
		seen[got] = true
	}
	if len(seen) < 2 {
		t.Errorf("Random() only produced %d distinct addresses", len(seen))
	}
}

// A full range must not report itself as empty.
func TestIPV4AddrRangeAddressCount(t *testing.T) {
	cases := []struct {
		from, to string
		want     int
	}{
		{"10.0.0.1", "10.0.0.1", 1},
		{"192.168.20.100", "192.168.20.150", 51},
		{"0.0.0.0", "255.255.255.255", 1 << 32},
	}
	for _, c := range cases {
		r := NewIPV4AddrRange(mustAddr(t, c.from), mustAddr(t, c.to))
		if got := r.AddressCount(); got != c.want {
			t.Errorf("AddressCount() over %s = %d, want %d", r, got, c.want)
		}
	}
}
