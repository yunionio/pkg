package netutils

import "testing"

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
