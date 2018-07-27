package secrules

import (
	"testing"
)

func TestIsFunction(t *testing.T) {
	cases := []struct {
		s          string
		s2         string
		bad        bool
		isAllowAny bool
	}{
		{s: "in:allow any", isAllowAny: true},
		{s: "out:allow any", isAllowAny: true},
		{s: "in:allow 0.0.0.0/0 any", s2: "in:allow any", isAllowAny: true},
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
		} else if c.isAllowAny != r.IsAllowAny() {
			t.Errorf("rule: %s: isAllowAny mismatch: want: %v, got: %v", c.s, c.isAllowAny, r.IsAllowAny())
		}
	}

}
