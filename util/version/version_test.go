package version

import (
	"testing"
)

func TestShortDate(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{
			in:   "2019-08-30T15:10:32Z",
			want: "19083015",
		},
		{
			in:   "20190830151032",
			want: "19083015",
		},
	}
	for _, c := range cases {
		got := shortDate(c.in)
		if got != c.want {
			t.Errorf("in: %s want: %s got: %s", c.in, c.want, got)
		}
	}
}
