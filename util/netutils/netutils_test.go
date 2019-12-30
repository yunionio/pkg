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
	"fmt"
	"testing"
)

func TestFormatMacAddr(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"92:10:31:a3:ec:37", "92:10:31:a3:ec:37"},
		{"92-10-31-A3-EC-37", "92:10:31:a3:ec:37"},
	}

	for _, c := range cases {
		if FormatMacAddr(c.in) != c.out {
			t.Errorf(" %s => %s != %s", c.in, FormatMacAddr(c.in), c.out)
		}
	}
}

func TestIP2Number(t *testing.T) {
	for _, addr := range []string{"192.168.23.1", "255.255.255.255", "0.0.0.0"} {
		num, err := IP2Number(addr)
		if err != nil {
			t.Errorf("IP2Number error %s %s", addr, err)
		}
		ipstr := Number2IP(num)
		if ipstr != addr {
			t.Errorf("%s != %s", ipstr, addr)
		}
	}
}

func TestIPV4Addr_StepDown(t *testing.T) {
	ipaddr, err := NewIPV4Addr("192.168.222.253")
	if err != nil {
		t.Errorf("NewIPV4Addr error %s", err)
	}
	t.Logf("Network: %s Broadcast: %s Client: %s", ipaddr.NetAddr(24), ipaddr.BroadcastAddr(24), ipaddr.CliAddr(24))
	t.Logf("Stepup: %s", ipaddr.StepUp())
	t.Logf("Stepdown: %s", ipaddr.StepDown())
}

func TestNewIPV4Addr(t *testing.T) {
	cases := []struct {
		in  string
		out IPV4Addr
	}{
		{
			in:  "", // maybe used by ":8080"
			out: IPV4Addr(0),
		},
		{
			in:  "0.0.0.0",
			out: IPV4Addr(0),
		},
		{
			in:  "192.168.1.0",
			out: IPV4Addr(192<<24 | 168<<16 | 1<<8),
		},
	}
	for _, c := range cases {
		got, err := NewIPV4Addr(c.in)
		if err != nil {
			t.Fatalf("(%q): err : %v", c.in, err)
		}
		if got != c.out {
			t.Fatalf("(%q): got %s, want %s", c.in, got, c.out)
		}
	}
}

func TestMasklen2Mask(t *testing.T) {
	t.Logf("%s", Masklen2Mask(0))
	t.Logf("%s", Masklen2Mask(1))
	t.Logf("%s", Masklen2Mask(23))
	t.Logf("%s", Masklen2Mask(24))
	t.Logf("%s", Masklen2Mask(32))

	t.Logf("%d", Mask2Len(Masklen2Mask(0)))
	t.Logf("%d", Mask2Len(Masklen2Mask(32)))
	t.Logf("%d", Mask2Len(Masklen2Mask(24)))
	t.Logf("%d", Mask2Len(Masklen2Mask(1)))
}

func TestIPRangeRandom(t *testing.T) {
	start, _ := NewIPV4Addr("192.168.20.100")
	end, _ := NewIPV4Addr("192.168.20.150")

	ipRange := NewIPV4AddrRange(end, start)

	for i := 0; i < 10; i += 1 {
		ip := ipRange.Random()
		t.Logf("%d: %s %s", i, ip, ip.ToMac("00:22:"))
	}
}

func TestIsExitAddress(t *testing.T) {
	for _, addr := range []string{"10.10.0.1", "172.31.32.1", "192.168.222.177", "114.113.226.53"} {
		ipv4, _ := NewIPV4Addr(addr)
		t.Logf("%s %v %v %v %v %v", ipv4.String(), IsPrivate(ipv4), IsHostLocal(ipv4), IsLinkLocal(ipv4), IsMulticast(ipv4), IsExitAddress(ipv4))
	}
}

func TestIPV4AddrRange_Contains(t *testing.T) {
	prefix, _ := NewIPV4Prefix("10.0.0.0/24")
	ipRange := prefix.ToIPRange()
	for _, addr := range []string{"10.0.0.24", "10.8.0.1"} {
		ipv4, _ := NewIPV4Addr(addr)
		t.Logf("%s contains %s %v", prefix.String(), ipv4, ipRange.Contains(ipv4))
	}
}

func TestIPV4AddrRange_Substract(t *testing.T) {
	nir := NewIPV4AddrRange
	ni := func(s string) IPV4Addr {
		i, err := NewIPV4Addr(s)
		if err != nil {
			msg := fmt.Sprintf("bad ip addr %q: %s", s, err)
			panic(msg)
		}
		return i
	}
	ar := nir(ni("192.168.2.0"), ni("192.168.2.255"))
	t.Run("disjoint (left)", func(t *testing.T) {
		ar2 := nir(ni("192.168.1.2"), ni("192.168.1.255"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 1 || !lefts[0].equals(ar) {
			t.Fatalf("bad `lefts`")
		}
		if sub != nil {
			t.Fatalf("bad `sub`: %#v", sub)
		}
	})
	t.Run("overlap (cut right)", func(t *testing.T) {
		ar2 := nir(ni("192.168.2.128"), ni("192.168.3.255"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 1 || !lefts[0].equals(nir(ni("192.168.2.0"), ni("192.168.2.127"))) {
			t.Fatalf("bad `lefts`")
		}
		if sub == nil || !sub.equals(nir(ni("192.168.2.128"), ni("192.168.2.255"))) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("contains (true subset)", func(t *testing.T) {
		ar2 := nir(ni("192.168.2.33"), ni("192.168.2.44"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 2 || !lefts[0].equals(nir(ni("192.168.2.0"), ni("192.168.2.32"))) || !lefts[1].equals(nir(ni("192.168.2.45"), ni("192.168.2.255"))) {
			t.Fatalf("bad `lefts`")
		}
		if sub == nil || !sub.equals(nir(ni("192.168.2.33"), ni("192.168.2.44"))) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("contains (align left)", func(t *testing.T) {
		ar2 := nir(ni("192.168.2.0"), ni("192.168.2.33"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 1 || !lefts[0].equals(nir(ni("192.168.2.34"), ni("192.168.2.255"))) {
			t.Fatalf("bad `lefts`")
		}
		if !sub.equals(nir(ni("192.168.2.0"), ni("192.168.2.33"))) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("contains (align right)", func(t *testing.T) {
		ar2 := nir(ni("192.168.2.44"), ni("192.168.2.255"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 1 || !lefts[0].equals(nir(ni("192.168.2.0"), ni("192.168.2.43"))) {
			t.Fatalf("bad `lefts`")
		}
		if !sub.equals(nir(ni("192.168.2.44"), ni("192.168.2.255"))) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("contained by", func(t *testing.T) {
		ar2 := nir(ni("192.168.1.255"), ni("192.168.3.0"))
		lefts, sub := ar.Substract(ar2)
		if len(lefts) != 0 {
			t.Fatalf("bad `lefts`")
		}
		if !sub.equals(ar) {
			t.Fatalf("bad `sub`")
		}
	})
}
