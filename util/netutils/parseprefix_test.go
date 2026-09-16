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

// Characterisation test: these are the spellings ParsePrefix currently
// accepts, including the ones that are more permissive than net.ParseCIDR.
// The expectations are pinned so that any future change to the grammar is a
// deliberate decision rather than an accident.
func TestParsePrefixAcceptedGrammar(t *testing.T) {
	cases := []struct {
		in       string
		wantAddr string
		wantLen  int8
	}{
		// Canonical forms.
		{"0.0.0.0/0", "0.0.0.0", 0},
		{"10.0.0.0/8", "10.0.0.0", 8},
		{"192.168.222.0/24", "192.168.222.0", 24},
		{"1.2.3.4/32", "1.2.3.4", 32},
		{"1.2.3.4/31", "1.2.3.4", 31},

		// A bare address is a /32.
		{"1.2.3.4", "1.2.3.4", 32},
		{"0.0.0.0", "0.0.0.0", 32},
		// An empty prefix is read as the zero address rather than rejected.
		{"", "0.0.0.0", 32},

		// Dotted-decimal masks, converted by counting leading one bits.
		{"10.0.0.0/255.0.0.0", "10.0.0.0", 8},
		{"1.2.3.4/255.255.255.0", "1.2.3.0", 24},
		{"1.2.3.4/255.255.255.255", "1.2.3.4", 32},
		// A non-contiguous mask is not rejected; it is read as its leading
		// run of one bits.
		{"1.2.3.4/255.0.255.0", "1.0.0.0", 8},
		// Leading zeros in the mask are read as zero bits.
		{"1.2.3.4/0.255.255.0", "0.0.0.0", 0},

		// The numeric form goes through strconv.Atoi.
		{"1.2.3.4/+8", "1.0.0.0", 8},
		{"1.2.3.4/024", "1.2.3.0", 24},
		{"1.2.3.4/08", "1.0.0.0", 8},

		// Leading zeros in the address octets are read as decimal digits.
		{"010.1.1.1/8", "10.0.0.0", 8},
	}
	for _, c := range cases {
		addr, maskLen, err := ParsePrefix(c.in)
		if err != nil {
			t.Errorf("ParsePrefix(%q): %v", c.in, err)
			continue
		}
		if addr.String() != c.wantAddr || maskLen != c.wantLen {
			t.Errorf("ParsePrefix(%q) = %s/%d, want %s/%d",
				c.in, addr, maskLen, c.wantAddr, c.wantLen)
		}
	}
}

// Values outside the accepted grammar are still rejected.
func TestParsePrefixRejected(t *testing.T) {
	cases := []string{
		"1.2.3/24",
		"1.2.3.4.5/24",
		"1.2.3.4/33",
		"1.2.3.4/-1",
		"1.2.3.4/abc",
		"1.2.3.4/",
		"300.1.1.1/24",
	}
	for _, in := range cases {
		if addr, maskLen, err := ParsePrefix(in); err == nil {
			t.Errorf("ParsePrefix(%q) returned no error, got %s/%d", in, addr, maskLen)
		}
	}
}

// The lenient forms are not accepted by the regular expressions that some
// callers use to validate a prefix first. Pinned so the difference between
// the two is visible.
func TestParsePrefixLenientFormsNotMatchedByRegutils(t *testing.T) {
	lenient := []string{
		"1.2.3.4/255.0.255.0", // dotted mask
		"1.2.3.4/0.255.255.0",
	}
	for _, in := range lenient {
		if _, _, err := ParsePrefix(in); err != nil {
			t.Errorf("ParsePrefix(%q) unexpectedly rejected", in)
		}
	}
}
