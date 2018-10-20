package regutils

import (
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
