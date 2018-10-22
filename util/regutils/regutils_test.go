package regutils

import (
	"fmt"
	"testing"
)

func TestIsFunction(t *testing.T) {
	cases := []struct {
		in  string
		out bool
	}{
		{"NOW()", true},
		{"test", false},
		{"CONCAT('123','456')", true},
	}
	for _, c := range cases {
		if MatchFunction(c.in) != c.out {
			t.Errorf("%s is a function: %v", c.in, c.out)
		}
	}

}

func TestRegexp(t *testing.T) {
	cases := []struct {
		in       string
		funcName string
		regfunc  func(str string) bool
		want     bool
	}{
		{"3.490.000,89", "MatchEUCurrency", MatchEUCurrency, true},
		{"3.000,89", "MatchEUCurrency", MatchEUCurrency, true},
		{"3.490", "MatchEUCurrency", MatchEUCurrency, true},
		{"0,89", "MatchEUCurrency", MatchEUCurrency, true},
		{"3.490.000,", "MatchEUCurrency", MatchEUCurrency, true},
		{"3,490,000,", "MatchEUCurrency", MatchEUCurrency, false},

		{"3,490,000,", "MatchUSCurrency", MatchUSCurrency, false},
		{"3,490,000.", "MatchUSCurrency", MatchUSCurrency, true},
		{"3,490,000.89", "MatchUSCurrency", MatchUSCurrency, true},
		{"3,000.89", "MatchUSCurrency", MatchUSCurrency, true},
		{"3,490", "MatchUSCurrency", MatchUSCurrency, true},
		{"0.89", "MatchUSCurrency", MatchUSCurrency, true},
		{"3.490.000,89", "MatchUSCurrency", MatchUSCurrency, false},
		{"3.000,89", "MatchUSCurrency", MatchUSCurrency, false},
		{"3.490", "MatchUSCurrency", MatchUSCurrency, true},
		{"0,89", "MatchUSCurrency", MatchUSCurrency, false},
		{"3.490.000,", "MatchUSCurrency", MatchUSCurrency, false},
		{"3.490.000.", "MatchEUCurrency", MatchEUCurrency, false},
		{",200.90", "MatchUSCurrency", MatchUSCurrency, false},
		{"1,200.90", "MatchUSCurrency", MatchUSCurrency, true},
		{".200,90", "MatchEUCurrency", MatchEUCurrency, false},
		{"1.200,90", "MatchEUCurrency", MatchEUCurrency, true},

		{"10.168.222.23", "MatchIPAddr", MatchIPAddr, true},
		{"10.168.222.", "MatchIPAddr", MatchIPAddr, false},

		{"1.0.168.192.in-addr.arpa", "MatchPtr", MatchPtr, true},
		{"0.168.192.in-addr.arpa", "MatchPtr", MatchPtr, false},
	}
	for _, c := range cases {
		got := c.regfunc(c.in)
		if got != c.want {
			t.Errorf("%s(%s) != %v", c.funcName, c.in, c.want)
		}
	}
}

func TestMatchIP4Addr(t *testing.T) {
	cases := []struct {
		in    string
		match bool
	}{
		{
			in:    "1.2.3.4",
			match: true,
		},
		{
			in:    "1.2.3.0",
			match: true,
		},
		{
			in:    "1.2.3.256",
			match: false,
		},
		{
			in:    "::1",
			match: false,
		},
		{
			in:    "::1.2.3.4",
			match: false,
		},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := MatchIP4Addr(c.in)
			if got != c.match {
				t.Errorf("%s: want %v, got %v", c.in, c.match, got)
			}
		})
	}
}

func TestMatchIP6Addr(t *testing.T) {
	cases := []struct {
		in    string
		match bool
	}{
		{
			in:    "",
			match: false,
		},
		{
			in:    "127.0.0.1",
			match: false,
		},
		{
			in:    "0.0.0.0",
			match: false,
		},
		{
			in:    "255.255.255.255",
			match: false,
		},
		{
			in:    "1.2.3.4",
			match: false,
		},
		{
			in:    "::1",
			match: true,
		},
		{
			in:    "2001:db8:0000:1:1:1:1:1",
			match: true,
		},
		{
			in:    "2001:db8:0000:1:1.2.3.4",
			match: false,
		},
		{
			in:    "2001:db8:0000:1::1.2.3.4",
			match: true,
		},
		{
			in:    "300.0.0.0",
			match: false,
		},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := MatchIP6Addr(c.in)
			if got != c.match {
				t.Errorf("%s: want %v, got %v", c.in, c.match, got)
			}
		})
	}
}

func TestMatchCIDR(t *testing.T) {
	cases := []struct {
		ip      string
		validIp bool
	}{
		{
			ip:      "1.2.3.4",
			validIp: true,
		},
		{
			ip:      "1.2.3.0",
			validIp: true,
		},
		{
			ip:      "1.2.3.256",
			validIp: false,
		},
		{
			ip:      "::1",
			validIp: false,
		},
		{
			ip:      "::1.2.3.4",
			validIp: false,
		},
	}
	for _, c := range cases {
		for _, m := range []int{-1, 0, 16, 32, 33} {
			in := fmt.Sprintf("%s/%d", c.ip, m)
			match := c.validIp && m >= 0 && m <= 32
			t.Run(in, func(t *testing.T) {
				got := MatchCIDR(in)
				if got != match {
					t.Errorf("%s: want %v, got %v", in, match, got)
				}
			})
		}
	}
}
