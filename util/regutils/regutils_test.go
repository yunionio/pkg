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
		{"May 6, 2019 07:01:07 AM", "MatchZStackTime", MatchZStackTime, true},
		{"2019-09-17T02:52:26.095915238+08:00", "MatchFullIsoNanotime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915238-08:00", "MatchFullIsoNanotime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915238Z", "MatchFullIsoNanotime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915Z", "MatchFullIsoTime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.09591Z", "MatchFullIsoTime2", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915+08:00", "MatchFullIsoTime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915-08:00", "MatchFullIsoTime", MatchFullISOTime, true},
		{"2019-09-17T02:52:26.095915+08:00", "MatchISOTime", MatchISOTime, false},
		{"2019-09-17T02:52:26+08:00", "MatchISOTime", MatchISOTime, true},
		{"2019-09-17T02:52:26-08:00", "MatchISOTime", MatchISOTime, true},
		{"2019-09-17 02:52:26-08:00", "MatchISOTime2", MatchISOTime2, true},
		{"2019-09-17 02:52:26.095915-08:00", "MatchFullIsoTime2", MatchFullISOTime2, true},
		{"2019-09-17 02:52-08:00", "MatchISONoSecondTime2", MatchISONoSecondTime2, true},
		{"2019-09-17T02:52-08:00", "MatchISONoSecondTime", MatchISONoSecondTime, true},
		{"12-31-22", "MatchDateExcel", MatchDateExcel, true},

		{"2022-06-01 23:59:59 +0000 UTC", "MatchClickhouseTime", MatchClickhouseTime, true},
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

func TestMatchDomainName(t *testing.T) {
	// Taken from https://github.com/asaskevich/govalidator
	var tests = []struct {
		param    string
		expected bool
	}{
		{"localhost", true},
		{"a.bc", true},
		{"a.b.", true},
		{"a.b..", false},
		{"localhost.local", true},
		{"localhost.localdomain.intern", true},
		{"l.local.intern", true},
		{"ru.link.n.svpncloud.com", true},
		{"-localhost", false},
		{"localhost.-localdomain", false},
		{"localhost.localdomain.-int", false},
		{"_localhost", true},
		{"localhost._localdomain", true},
		{"localhost.localdomain._int", true},
		{"lÖcalhost", false},
		{"localhost.lÖcaldomain", false},
		{"localhost.localdomain.üntern", false},
		{"__", true},
		{"localhost/", false},
		{"127.0.0.1", false},
		{"[::1]", false},
		{"50.50.50.50", false},
		{"localhost.localdomain.intern:65535", false},
		{"漢字汉字", false},
		{"www.jubfvq1v3p38i51622y0dvmdk1mymowjyeu26gbtw9andgynj1gg8z3msb1kl5z6906k846pj3sulm4kiyk82ln5teqj9nsht59opr0cs5ssltx78lfyvml19lfq1wp4usbl0o36cmiykch1vywbttcus1p9yu0669h8fj4ll7a6bmop505908s1m83q2ec2qr9nbvql2589adma3xsq2o38os2z3dmfh2tth4is4ixyfasasasefqwe4t2ub2fz1rme.de", false},
	}

	for _, test := range tests {
		actual := MatchDomainName(test.param)
		if actual != test.expected {
			t.Errorf("Expected MatchDomainName(%q) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestMatchDomainSRV(t *testing.T) {
	// Taken from https://github.com/asaskevich/govalidator
	var tests = []struct {
		param    string
		expected bool
	}{
		{"_etcd._tcp.example.com", true},
		{"_._", false},
		{"_._.", false},
		{"_a._.", false},
		{"_a._b.", false},
		{"_a._b.c", true},
		{"a", false},
		{"a.b", false},
		{"a.b.c", false},
		{"_a.b.c", false},
	}

	for _, test := range tests {
		actual := MatchDomainSRV(test.param)
		if actual != test.expected {
			t.Errorf("Expected MatchDomainSRV(%q) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}
