package utils

import "testing"

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
