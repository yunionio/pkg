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

package secrules

import (
	"sort"
	"testing"

	"yunion.io/x/log"
	"yunion.io/x/pkg/util/netutils"
)

func TestSecRuleSet_AllowList(t *testing.T) {
	srs0 := SecurityRuleSet{
		*MustParseSecurityRule("out:deny 192.168.222.2 tcp 3389"),
		*MustParseSecurityRule("out:allow any"),
	}
	rules := srs0.AllowList()
	a, _ := netutils.NewIPV4Addr("192.168.222.2")
	for _, rule := range rules {
		switch rule.Protocol {
		case PROTO_TCP:
			ar := netutils.NewIPV4AddrRangeFromIPNet(rule.IPNet)
			if ar.Contains(a) && rule.PortStart <= 3389 && rule.PortEnd >= 3389 {
				log.Fatalf("allow list should not contain 192.168.222.2 tcp 3389")
			}
		case PROTO_ICMP, PROTO_UDP:
			if rule.IPNet.String() != "0.0.0.0/0" {
				log.Fatalf("proto %s shoud be merged", rule.Protocol)
			}
		}
	}
	dieIf := func(t *testing.T, srs0, srs1 SecurityRuleSet) {
		sort.Sort(srs0)
		sort.Sort(srs1)
		if !srs0.equals(srs1) {
			t.Fatalf("not equal:\n%s\n%s", srs0, srs1)
		}
	}
	t.Run("empty", func(t *testing.T) {
		srs0 := SecurityRuleSet{}
		srs1 := srs0.AllowList()
		dieIf(t, srs0, srs1)
	})
	t.Run("all allow", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:allow any"),
		}
		srs1 := srs0.AllowList()
		dieIf(t, srs0, srs1)
	})
	t.Run("annihilate: reduce to nothing", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:deny any"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 any"),
			*MustParseSecurityRule("in:allow 0.0.0.0/0 tcp"),
			*MustParseSecurityRule("in:allow 0.0.0.0/0 icmp"),
			*MustParseSecurityRule("in:allow 8.0.0.0/0 tcp 3,4"),
			*MustParseSecurityRule("in:allow 8.0.0.0/0 udp 3,4"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{}
		dieIf(t, srs1, srs1_)
	})
	t.Run("net: allow;deny;allow", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/25 any"),
			*MustParseSecurityRule("in:deny 192.168.2.0/24 any"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 any"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/25 any"),
			*MustParseSecurityRule("in:allow 192.168.3.0/24 any"),
		}
		dieIf(t, srs1, srs1_)
	})

	t.Run("net: tick out singles", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:deny 192.168.2.33 any"),
			*MustParseSecurityRule("in:deny 192.168.1.8 any"),
			*MustParseSecurityRule("in:deny 192.168.33.8 any"),
			*MustParseSecurityRule("in:allow 192.168.2.0/24 any"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/27 any"),
			*MustParseSecurityRule("in:allow 192.168.2.128/25 any"),
			*MustParseSecurityRule("in:allow 192.168.2.32 any"),
			*MustParseSecurityRule("in:allow 192.168.2.34/31 any"),
			*MustParseSecurityRule("in:allow 192.168.2.36/30 any"),
			*MustParseSecurityRule("in:allow 192.168.2.40/29 any"),
			*MustParseSecurityRule("in:allow 192.168.2.48/28 any"),
			*MustParseSecurityRule("in:allow 192.168.2.64/26 any"),
		}
		{
			a, _ := netutils.NewIPV4Addr("192.168.2.33")
			sum := 0
			for _, r := range srs1_ {
				ar := netutils.NewIPV4AddrRangeFromIPNet(r.IPNet)
				sum += ar.AddressCount()
				t.Logf("left range: %s", ar.String())
				if ar.Contains(a) {
					t.Fatalf("  contains %s", a.String())
				}
			}
			if sum != 255 {
				t.Fatalf("expecting a total of 256-1 addresses, got: %d", sum)
			}
		}
		dieIf(t, srs1, srs1_)
	})

	t.Run("port range: deny tcp", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:deny tcp 1-1024"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 any"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/23 icmp"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 tcp 1025-65535"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 udp 1-65535"),
		}
		dieIf(t, srs1, srs1_)
	})
	t.Run("port range: deny tcp&udp same range", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:deny tcp 1-1024"),
			*MustParseSecurityRule("in:deny udp 1-1024"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 any"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/23 icmp"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 tcp 1025-65535"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 udp 1025-65535"),
		}
		dieIf(t, srs1, srs1_)
	})
	t.Run("port range: deny tcp&udp diff range", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:deny tcp 1-1024"),
			*MustParseSecurityRule("in:deny udp 22-1024"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 any"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/23 icmp"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 tcp 1025-65535"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 udp 1-21"),
			*MustParseSecurityRule("in:allow 192.168.2.0/23 udp 1025-65535"),
		}
		dieIf(t, srs1, srs1_)
	})
	t.Run("ports: cannot merge", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/24 tcp 22,80"),
			*MustParseSecurityRule("in:allow 192.168.3.0/24 tcp 8080,3389"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/24 tcp 22,80"),
			*MustParseSecurityRule("in:allow 192.168.3.0/24 tcp 3389,8080"),
		}
		dieIf(t, srs1, srs1_)
	})
	t.Run("ports: merge", func(t *testing.T) {
		srs0 := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/24 tcp 22,80"),
			*MustParseSecurityRule("in:allow 192.168.2.0/24 tcp 8080,3389"),
		}
		srs1 := srs0.AllowList()
		srs1_ := SecurityRuleSet{
			*MustParseSecurityRule("in:allow 192.168.2.0/24 tcp 22,80,3389,8080"),
		}
		dieIf(t, srs1, srs1_)
	})
}
