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

import "testing"

// An IPv4-mapped IPv6 CIDR normalises to IPv4. The prefix length that is
// printed has to be the one of the normalised form, otherwise the text no
// longer describes the range the rule covers.
func TestSecurityRuleStringKeepsPrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"in:allow ::ffff:0:0/96 tcp 22", "in:allow 0.0.0.0/0 tcp 22"},
		{"out:deny ::ffff:0:0/96 any", "out:deny 0.0.0.0/0 any"},
		{"in:allow ::ffff:0102:0300/120 tcp 22", "in:allow 1.2.3.0/24 tcp 22"},
	}
	for _, c := range cases {
		rule, err := ParseSecurityRule(c.in)
		if err != nil {
			t.Errorf("ParseSecurityRule(%q): %v", c.in, err)
			continue
		}
		if got := rule.String(); got != c.want {
			t.Errorf("ParseSecurityRule(%q).String() = %q, want %q", c.in, got, c.want)
		}
	}
}

// The text form has to survive a parse/print round trip without changing the
// range the rule covers.
func TestSecurityRuleStringRoundTrip(t *testing.T) {
	inputs := []string{
		"in:allow ::ffff:0:0/96 tcp 22",
		"out:deny ::ffff:0:0/96 any",
		"in:allow ::ffff:0102:0300/120 tcp 22",
		"in:allow 0.0.0.0/0 any",
		"in:allow 10.0.8.0/24 tcp 80",
		"in:allow fd:3ffe:3200:1220::/64 any",
		"in:allow 0.0.0.0 tcp",
		"in:allow :: tcp",
	}
	for _, in := range inputs {
		first, err := ParseSecurityRule(in)
		if err != nil {
			t.Errorf("ParseSecurityRule(%q): %v", in, err)
			continue
		}
		text := first.String()

		second, err := ParseSecurityRule(text)
		if err != nil {
			t.Errorf("re-parsing %q: %v", text, err)
			continue
		}
		if second.String() != text {
			t.Errorf("%q is not stable: first %q, second %q", in, text, second.String())
		}
		if !first.equals(second) {
			t.Errorf("%q: reparsed rule is not equal to the original", in)
		}
	}
}

// Rules that were already printed with a prefix keep the same text.
func TestSecurityRuleStringUnchangedForNormalForms(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"in:allow 0.0.0.0/0 any", "in:allow 0.0.0.0/0 any"},
		{"in:allow ::/0 any", "in:allow ::/0 any"},
		{"in:allow 10.0.8.0/24 tcp 80", "in:allow 10.0.8.0/24 tcp 80"},
		{"in:allow fd:3ffe:3200:1220::/64 any", "in:allow fd:3ffe:3200:1220::/64 any"},
		{"in:allow 0.0.0.0 tcp", "in:allow 0.0.0.0 tcp"},
		{"in:allow :: tcp", "in:allow :: tcp"},
	}
	for _, c := range cases {
		rule, err := ParseSecurityRule(c.in)
		if err != nil {
			t.Errorf("ParseSecurityRule(%q): %v", c.in, err)
			continue
		}
		if got := rule.String(); got != c.want {
			t.Errorf("ParseSecurityRule(%q).String() = %q, want %q", c.in, got, c.want)
		}
	}
}
