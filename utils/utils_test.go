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
