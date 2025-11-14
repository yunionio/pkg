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

package utils

import (
	"reflect"
	"testing"
)

func TestUnquote(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`abc`, `abc`},
		{`"abc"`, `abc`},
		{`"'\"abc\"'"`, `'"abc"'`},
		{`'"\'abc\'"'`, `"'abc'"`},
		{`hello\nworld`, `hello\nworld`},
		{`"hello\nworld"`, "hello\nworld"},
		{`'hello\nworld'`, "hello\nworld"},
		{"hello\nworld", "hello"},
		{`"\thello\n\rworld"` + "ether\nether", "\thello\n\rworld"},
	}
	for _, c := range cases {
		if got := Unquote(c.in); got != c.want {
			t.Errorf("Unquote(%q) got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCamel2Kebab(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"TEST", "test"},
		{"GPUTestScore", "gpu-test-score"},
		{"UserName", "user-name"},
		{"AuthURL", "auth-url"},
		{"Auth_URL", "auth-url"},
		{"Ip6Addr", "ip6-addr"},
		{"SourceVSwitchId", "source-vswitch-id"},
		{"SNATEntry", "snat-entry"},
		{"HAProxy", "ha-proxy"},
		{"ATest", "atest"},
		{"SAMLMetadata", "saml-metadata"},
		{"IPAddress", "ip-address"},
	}
	for _, c := range cases {
		got := CamelSplit(c.in, "-")
		if got != c.want {
			t.Errorf("Camel2Kebab(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestKebab2Camel(t *testing.T) {
	cases := []struct {
		in, sep, want string
	}{
		{"test", "-", "Test"},
		{"user-name", "-", "UserName"},
		{"auth-url", "-", "AuthUrl"},
		{"on_init", "_", "OnInit"},
	}
	for _, c := range cases {
		got := Kebab2Camel(c.in, c.sep)
		if got != c.want {
			t.Errorf("Kebab2Camel(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestFloatRound(t *testing.T) {
	type args struct {
		num       float64
		precision int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "test 0.00",
			args: args{0, 2},
			want: 0.00,
		},
		{
			name: "test 12.12345",
			args: args{12.12645, 2},
			want: 12.120000000000001,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatRound(tt.args.num, tt.args.precision); got != tt.want {
				t.Errorf("FloatRound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArgsStringToArray(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test server-list   --details",
			args: args{"server-list   --details"},
			want: []string{"server-list", "--details"},
		},
		{
			name: "test server-monitor  id \"info block\"",
			args: args{"server-monitor  id \"info block\""},
			want: []string{"server-monitor", "id", "info block"},
		},
		{
			name: "test x 'aa'bb bb'aa' aa'bb'cc   dd\"aa\"cc'bb'ee   ",
			args: args{"x 'aa'bb bb'aa' aa'bb'cc   dd\"aa\"cc'bb'ee   "},
			want: []string{"x", "aabb", "bbaa", "aabbcc", "ddaaccbbee"},
		},
		{
			name: "test 0 '1'2\"3 4\"5 6",
			args: args{"0 '1'2\"3 4\"5 6"},
			want: []string{"0", "123 45", "6"},
		},
		{
			name: "test abc\"'\"$'\"\"\"",
			args: args{"abc\"'\"$\"\"\""},
			want: []string{"abc'$\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ArgsStringToArray(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArgsStringToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	cases := []struct {
		In    string
		Trunc int
		Want  string
	}{
		{"abc", 5, "abc"},
		{"abc", 2, "ab.."},
	}
	for _, c := range cases {
		got := TruncateString(c.In, c.Trunc)
		if got != c.Want {
			t.Errorf("In: %s Trunc: %d Want: %s Got: %s", c.In, c.Trunc, c.Want, got)
		}
	}
}

func BenchmarkCamelSplitTokens(b *testing.B) {
	cases := []struct {
		in   string
		want string
	}{
		{"TEST", "test"},
		{"GPUTestScore", "gpu-test-score"},
		{"UserName", "user-name"},
		{"AuthURL", "auth-url"},
	}
	for _, c := range cases {
		b.Run(c.in, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				CamelSplitTokens(c.in)
			}
		})
	}
}

func TestInArray(t *testing.T) {
	cases := []struct {
		needle int
		array  []int
		want   bool
	}{
		{
			needle: 2,
			array:  []int{1, 2},
			want:   true,
		},
		{
			needle: 3,
			array:  []int{1, 2},
			want:   false,
		},
	}
	for _, c := range cases {
		got := IsInArray(c.needle, c.array)
		if got != c.want {
			t.Errorf("got: %v want %v", got, c.want)
		}
	}
}

func TestCamelSplit(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"TEST", "test"},
		{"GPUTestScore", "gpu_test_score"},
		{"UserName", "user_name"},
		{"AuthURL", "auth_url"},
		{"ID_", "id_"},
		{"DEPLOYMENT_ID_", "deployment_id_"},
	}
	for _, c := range cases {
		got := CamelSplit(c.in, "_")
		if got != c.want {
			t.Errorf("CamelSplit(%s) = %s, want %s", c.in, got, c.want)
		}
	}
}
