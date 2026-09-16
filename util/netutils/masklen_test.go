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

func TestMasklen2MaskInRange(t *testing.T) {
	cases := []struct {
		maskLen int8
		want    uint32
	}{
		{0, 0x00000000},
		{8, 0xff000000},
		{24, 0xffffff00},
		{31, 0xfffffffe},
		{32, 0xffffffff},
	}
	for _, c := range cases {
		if got := uint32(Masklen2Mask(c.maskLen)); got != c.want {
			t.Errorf("Masklen2Mask(%d) = %#08x, want %#08x", c.maskLen, got, c.want)
		}
	}
}

// A length above 32 is not representable. It must not wrap around into a
// match-all mask, which would widen the prefix instead of narrowing it.
func TestMasklen2MaskAboveRange(t *testing.T) {
	for _, maskLen := range []int8{33, 40, 64, 100, 127} {
		got := Masklen2Mask(maskLen)
		if uint32(got) != 0xffffffff {
			t.Errorf("Masklen2Mask(%d) = %#08x, want %#08x", maskLen, uint32(got), uint32(0xffffffff))
		}
		if l := Mask2Len(got); l != 32 {
			t.Errorf("Mask2Len(Masklen2Mask(%d)) = %d, want 32", maskLen, l)
		}
	}
}

func TestNewIPV4PrefixFromAddrClampsAboveRange(t *testing.T) {
	addr, err := NewIPV4Addr("1.2.3.4")
	if err != nil {
		t.Fatalf("NewIPV4Addr: %v", err)
	}
	for _, masklen := range []int8{33, 40, 127} {
		pref := NewIPV4PrefixFromAddr(addr, masklen)
		if pref.MaskLen != 32 {
			t.Errorf("NewIPV4PrefixFromAddr(1.2.3.4, %d).MaskLen = %d, want 32", masklen, pref.MaskLen)
		}
		r := pref.ToIPRange()
		if r.StartIp() != addr || r.EndIp() != addr {
			t.Errorf("NewIPV4PrefixFromAddr(1.2.3.4, %d) covers %s, want just 1.2.3.4", masklen, r)
		}
		if r.AddressCount() != 1 {
			t.Errorf("NewIPV4PrefixFromAddr(1.2.3.4, %d) covers %d addresses, want 1", masklen, r.AddressCount())
		}
	}
}

// Lengths inside the range must produce exactly the same prefixes as before.
func TestNewIPV4PrefixFromAddrInRange(t *testing.T) {
	cases := []struct {
		masklen   int8
		net       string
		broadcast string
	}{
		{0, "0.0.0.0", "255.255.255.255"},
		{8, "1.0.0.0", "1.255.255.255"},
		{24, "1.2.3.0", "1.2.3.255"},
		{32, "1.2.3.4", "1.2.3.4"},
	}
	addr, err := NewIPV4Addr("1.2.3.4")
	if err != nil {
		t.Fatalf("NewIPV4Addr: %v", err)
	}
	for _, c := range cases {
		pref := NewIPV4PrefixFromAddr(addr, c.masklen)
		if pref.MaskLen != c.masklen {
			t.Errorf("MaskLen = %d, want %d", pref.MaskLen, c.masklen)
		}
		r := pref.ToIPRange()
		if r.StartIp().String() != c.net {
			t.Errorf("masklen %d starts at %s, want %s", c.masklen, r.StartIp(), c.net)
		}
		if r.EndIp().String() != c.broadcast {
			t.Errorf("masklen %d ends at %s, want %s", c.masklen, r.EndIp(), c.broadcast)
		}
	}
}
