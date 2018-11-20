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
		{[]string{"A", "B", "C"},
			[][]string{
				{"a", "云联 华北2（北京）\t\t", "ok"},
				{"b", "\t云联 印度尼西亚（雅加达）", "ok"},
				{"c", "云联 香港", "ok"},
			},
			`+---+------------------------------------+----+
| A |                 B                  | C  |
+---+------------------------------------+----+
| a | 云联 华北2（北京）		 | ok |
| b | 	云联 印度尼西亚（雅加达）        | ok |
| c | 云联 香港                          | ok |
+---+------------------------------------+----+
`,
		},
	}
	for _, c := range cases {
		pt := NewPrettyTable(c.title)
		out := pt.GetString(c.in)
		if out != c.out {
			t.Errorf("got != want\n%s\n(%d)\n !=\n%s\n(%d)\n", out, len(out), c.out, len(c.out))
		}
	}
}
