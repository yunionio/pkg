package secrules

import (
	"testing"
)

func TestIsFunction(t *testing.T) {
	cases := []struct {
		s           string
		s2          string
		bad         bool
		isWildMatch bool
	}{
		{s: "in:allow any", isWildMatch: true},
		{s: "out:allow any", isWildMatch: true},
		{s: "in:deny any", isWildMatch: true},
		{s: "out:deny any", isWildMatch: true},
		{s: "in:allow 0.0.0.0/0 any", s2: "in:allow any", isWildMatch: true},
		{s: "in:allow 0.0.0.0/0 tcp", s2: "in:allow tcp"},
		{s: "in:allow 0.0.0.0/0 udp", s2: "in:allow udp"},
		{s: "in:allow 0.0.0.0/0 icmp", s2: "in:allow icmp"},
		{s: "in:allow 10.0.8.0/24 any", s2: "in:allow 10.0.8.0/24 any"},
		{s: "in:allow 10.0.9.0/24 tcp", s2: "in:allow 10.0.9.0/24 tcp"},
		{s: "in:allow 10.0.10.0/24 udp", s2: "in:allow 10.0.10.0/24 udp"},
		{s: "in:allow 10.0.11.0/24 icmp", s2: "in:allow 10.0.11.0/24 icmp"},
		{s: "in:allow 10.0.8.0/24 tcp 1-100", s2: "in:allow 10.0.8.0/24 tcp 1-100"},
		{s: "in:allow 10.0.8.0/24 tcp 100-1", s2: "in:allow 10.0.8.0/24 tcp 1-100"},
		{s: "in:allow 10.0.8.0/24 tcp 1,100", s2: "in:allow 10.0.8.0/24 tcp 1,100"},
		{s: "in:allow 10.0.8.0/24 tcp 100", s2: "in:allow 10.0.8.0/24 tcp 100"},
		{s: "in:allow 0.0.0.0 tcp", s2: "in:allow 0.0.0.0 tcp"},
		{s: "in:allow 0.0.0.0 tcp", s2: "in:allow 0.0.0.0 tcp"},
		{s: "in:deny", bad: true},
		{s: "in:allow", bad: true},
		{s: "in:allow 0.0.0.0/0 ip", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp 0", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp -1", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp 65536", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp 0,100", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp -1,100", bad: true},
		{s: "in:allow 10.0.8.0/24 tcp 10--1", bad: true},
	}
	for _, c := range cases {
		r, err := ParseSecurityRule(c.s)
		if err != nil {
			if !c.bad {
				t.Errorf("rule: %s: parse error: %s", c.s, err)
			}
			if r != nil {
				t.Errorf("rule: %s: expect nil on parse error: %s", c.s, err)
			}
			continue
		} else if c.bad {
			t.Errorf("rule: %s: bad but parsed: %s", c.s, r.String())
			continue
		} else if len(c.s2) > 0 && c.s2 != r.String() {
			t.Errorf("rule: %s: parsed but wrong:\n\twant: %s,\n\tgot: %s",
				c.s, c.s2, r.String())
		} else if c.isWildMatch != r.IsWildMatch() {
			t.Errorf("rule: %s: wildMatch mismatch: want: %v, got: %v",
				c.s, c.isWildMatch, r.IsWildMatch())
		}
	}

}

func TestValidateRuleFunction(t *testing.T) {
	cases := []struct {
		rule SecurityRule
		bad  bool
	}{
		{
			rule: SecurityRule{},
			bad:  true,
		},
		{
			rule: SecurityRule{
				Direction: "hello",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "world",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "test",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  -1,
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  0,
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  1,
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  101,
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				PortStart: -1,
				PortEnd:   -1,
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				PortStart: 10,
				PortEnd:   7,
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				PortStart: 10,
				PortEnd:   11,
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				PortStart: 10,
				PortEnd:   10,
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				PortStart: 10,
				PortEnd:   1000000,
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				Ports:     []int{},
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				Ports:     []int{1, 2, 3},
			},
			bad: false,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				Ports:     []int{1, 2, 3, 0},
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "any",
				Priority:  100,
				Ports:     []int{1, 2, 3, 23232323},
			},
			bad: true,
		},
		{
			rule: SecurityRule{
				Direction: "in",
				Action:    "allow",
				Protocol:  "icmp",
				Priority:  100,
				Ports:     []int{1, 2, 3, 23232323},
			},
			bad: true,
		},
	}

	for _, c := range cases {
		err := c.rule.ValidateRule()
		if err != nil {
			if !c.bad {
				t.Errorf("rule: %v validate error: %v", c.rule, err)
			}
		} else if c.bad {
			t.Errorf("rule: %v bad but validate pass", c.rule)
		}
	}

}
