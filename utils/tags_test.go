package utils

import (
	"reflect"
	"testing"
)

func TestFindWord(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{`'abc'`, `abc`},
		{`"abc"`, `abc`},
		{`'id.in(123-123,456-456)'`, `id.in(123-123,456-456)`},
		{`--config`, `--config`},
	}
	for _, c := range cases {
		o := Unquote(c.in)
		t.Logf("in: %s out: %s expect: %s", c.in, o, c.out)
	}
}

func TestFindWords(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		want  []string
		panic bool
	}{
		{
			name: "double quoted",
			in:   `"2018-08-31 15:20:33"`,
			want: []string{`2018-08-31 15:20:33`},
		},
		{
			name: "single quoted",
			in:   `'2018-08-31 15:20:33'`,
			want: []string{`2018-08-31 15:20:33`},
		},
		{
			name:  "panic",
			in:    `2018-08-31 15:20:33`,
			panic: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				v := recover()
				if v != nil {
					if !c.panic {
						t.Fatalf("panic: %s", v)
					}
				} else {
					if c.panic {
						t.Fatalf("want panic, but did not happen")
					}
				}
			}()
			got := FindWords([]byte(c.in), 0)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("want %#v, got %#v", c.want, got)
			}
		})
	}
}
