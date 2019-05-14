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
