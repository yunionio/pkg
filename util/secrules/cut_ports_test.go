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
)

// collectPortRanges flattens the port ranges of the cuts that were not
// marked as cut, i.e. the ports that remain allowed.
func collectPortRanges(cuts securityRuleCuts) [][2]int {
	ranges := [][2]int{}
	for _, c := range cuts {
		if c.portCut {
			continue
		}
		ranges = append(ranges, [2]int{c.r.PortStart, c.r.PortEnd})
	}
	sort.Slice(ranges, func(i, j int) bool { return ranges[i][0] < ranges[j][0] })
	return ranges
}

func portsAllowed(cuts securityRuleCuts) map[int]bool {
	allowed := map[int]bool{}
	for _, r := range collectPortRanges(cuts) {
		for p := r[0]; p <= r[1]; p++ {
			allowed[p] = true
		}
	}
	return allowed
}

// newAnyPortCut builds a rule that matches every port without an explicit
// range or port list, which is the shape that reaches the port-list path.
func newAnyPortCut() securityRuleCuts {
	return securityRuleCuts{
		{
			r: SecurityRule{
				Protocol: "tcp",
			},
		},
	}
}

// Cutting a single low port must not hand that port back through the
// trailing "everything above" range.
func TestCutOutPortsSingle(t *testing.T) {
	cases := []struct {
		cut     []uint16
		blocked []int
	}{
		{[]uint16{1}, []int{1}},
		{[]uint16{2}, []int{2}},
		{[]uint16{3}, []int{3}},
		{[]uint16{80}, []int{80}},
		{[]uint16{65535}, []int{65535}},
		{[]uint16{1, 3}, []int{1, 3}},
		{[]uint16{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5}},
		{[]uint16{1, 65535}, []int{1, 65535}},
	}
	for _, c := range cases {
		allowed := portsAllowed(newAnyPortCut().cutOutPorts("tcp", c.cut))
		for _, p := range c.blocked {
			if allowed[p] {
				t.Errorf("cutOutPorts(%v): port %d is still allowed", c.cut, p)
			}
		}
		// A port well away from the cut list stays allowed.
		if !allowed[40000] {
			t.Errorf("cutOutPorts(%v): port 40000 should still be allowed", c.cut)
		}
	}
}

// Cutting the whole range leaves nothing allowed.
func TestCutOutPortsEverything(t *testing.T) {
	all := make([]uint16, 0, 65535)
	for p := 1; p <= 65535; p++ {
		all = append(all, uint16(p))
	}
	if allowed := portsAllowed(newAnyPortCut().cutOutPorts("tcp", all)); len(allowed) != 0 {
		t.Errorf("cutting every port left %d ports allowed", len(allowed))
	}
}

// An empty cut list leaves the range untouched.
func TestCutOutPortsNone(t *testing.T) {
	ranges := collectPortRanges(newAnyPortCut().cutOutPorts("tcp", nil))
	if len(ranges) != 1 || ranges[0] != [2]int{1, 65535} {
		t.Errorf("cutOutPorts(nil) = %v, want [[1 65535]]", ranges)
	}
}

// A rule for a different protocol is passed through untouched, keeping the
// port values it was built with.
func TestCutOutPortsOtherProtocol(t *testing.T) {
	cuts := newAnyPortCut().cutOutPorts("udp", []uint16{1})
	ranges := collectPortRanges(cuts)
	if len(ranges) != 1 || ranges[0] != [2]int{0, 0} {
		t.Errorf("cutOutPorts for udp = %v, want [[0 0]]", ranges)
	}
}
