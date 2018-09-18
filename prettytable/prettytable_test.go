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
		{[]string{"A", "B"}, [][]string{}, ""},
		{[]string{"A", "B"},
			[][]string{
				{"a", "b"},
				{"c", "d"},
				{"e", "e\n"},
				{"f", "f\nf"},
				{"g", "g\ng\n"},
			},
			`+---+---+
| A | B |
+---+---+
| a | b |
| c | d |
| e | e |
| f | f |
|   | f |
| g | g |
|   | g |
+---+---+
`,
		},
		{[]string{"A", "B"},
			[][]string{
				{"a", "a\t\t"},
				{"b", "\tb\t"},
				{"c", "\t\tc"},
			},
			`+---+-------------+
| A |      B      |
+---+-------------+
| a | a		  |
| b | 	b	  |
| c | 		c |
+---+-------------+
`,
		},
	}
	for _, c := range cases {
		pt := NewPrettyTable(c.title)
		out := pt.GetString(c.in)
		if out != c.out {
			t.Errorf("\n%s\n(%d)\n !=\n%s\n(%d)\n", out, len(out), c.out, len(c.out))
		}
	}
}
