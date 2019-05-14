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
	"strings"
	"testing"
)

func TestRules(t *testing.T) {
	rules := []string{"in:allow 192.168.0.1/32 tcp 80", "out:deny any", "in:allow any", "out:allow udp 3232-3000", "in:allow tcp"}
	for _, r := range rules {
		rule, err := ParseSecurityRule(r)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(rule)
	}

	ruleSet := [][]string{
		[]string{"in:allow any"},
		[]string{"in:allow any", "in:deny tcp 25", "in:deny tcp 80"},
		[]string{"in:allow tcp 80", "in:allow udp"},
		[]string{"in:allow any", "in:allow tcp 80"},
		[]string{"in:allow tcp 1000-1999", "in:allow tcp 2000", "in:allow tcp 2200"},
		[]string{"in:allow tcp 1000-1999", "in:allow tcp 2000", "in:allow tcp 2200",
			"in:deny tcp 500-1500", "in:deny tcp 2000-3000", "in:deny tcp 1888"},
		[]string{"in:allow 192.168.0.0/16 tcp", "in:allow tcp 80",
			"out:deny 192.168.0.0/16 tcp 80", "out:deny tcp 80", "in:allow any",
			"in:allow 192.168.0.0/24 tcp", "out:deny 192.168.0.0/24 udp"},
	}

	for _, rs := range ruleSet {
		srs := SecurityGroupRuleSet{}
		for _, r := range rs {
			rule, err := ParseSecurityRule(r)
			if err != nil {
				t.Fatalf("parse rule %s error: %v", r, err)
			}
			srs.AddRule(*rule)
		}
		t.Log(srs.String())
	}

	ruleCompareSet := map[string][2][]string{
		"test_equal": {
			[]string{"in:allow any", "in:allow tcp"},
			[]string{"in:allow any"},
		},
		"test1_notequal": {
			[]string{"in:allow any", "out:allow tcp"},
			[]string{"in:allow any"},
		},
		"test2_equal": {
			[]string{"in:deny tcp 90-120", "in:deny tcp 120-200"},
			[]string{"in:deny tcp 90-200"},
		},
		"test3_equal": {
			[]string{"in:allow tcp 3389", "in:allow tcp 22"},
			[]string{"in:allow tcp 22", "in:allow tcp 3389"},
		},
		"test4_equal": {
			[]string{"in:allow any", "out:allow any"},
			[]string{"out:allow any", "in:allow any"},
		},
		"test4_notequal": {
			[]string{"in:allow tcp 3389", "in:allow tcp 22", "out:allow any"},
			[]string{"out:allow any"},
		},
		"test5_notequal": {
			[]string{"in:allow any"},
			[]string{"out:allow any"},
		},
	}
	for name, rs := range ruleCompareSet {
		rs0 := SecurityGroupRuleSet{}
		for _, r := range rs[0] {
			rule, err := ParseSecurityRule(r)
			if err != nil {
				t.Fatalf("parse rule %s error: %v", r, err)
			}
			rs0.AddRule(*rule)
		}
		rs1 := SecurityGroupRuleSet{}
		for _, r := range rs[1] {
			rule, err := ParseSecurityRule(r)
			if err != nil {
				t.Fatalf("parse rule %s error: %v", r, err)
			}
			rs1.AddRule(*rule)
		}
		t.Logf("rs0: %s rs1: %s", rs0.String(), rs1.String())
		equal := rs0.IsEqual(rs1)
		if strings.HasSuffix(name, "_equal") {
			if !equal {
				t.Fatalf("test %s failed should equal", name)
			}
		} else {
			if equal {
				t.Fatalf("test %s failed should not equal", name)
			}
		}
		t.Logf("test %s pass", name)
	}
}
