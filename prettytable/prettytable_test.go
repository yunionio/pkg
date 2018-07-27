package prettytable

import (
	"testing"
)

func TestGetString(t *testing.T) {
	cases := []struct {
		title []string
		in    [][]string
		out   string
	}{
		{[]string{"A", "B"},
			[][]string{
				{"a", "b"},
				{"c", "d"},
			},
			`+---+---+
| A | B |
+---+---+
| a | b |
| c | d |
+---+---+`,
		},
		{[]string{"A", "B"}, [][]string{}, ""},
	}
	for _, c := range cases {
		pt := NewPrettyTable(c.title)
		out := pt.GetString(c.in)
		if out != c.out {
			t.Errorf("\n%s\n(%d)\n !=\n%s\n(%d)\n", out, len(out), c.out, len(c.out))
		}
	}
}
