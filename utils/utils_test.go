package utils

import (
	"testing"
)

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
