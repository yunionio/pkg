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
		maxLineWidth int
		title        []string
		in           [][]string
		out          string
	}{
		{
			title: []string{"A", "B"},
			in:    [][]string{},
			out:   "",
		},
		{
			title: []string{"A", "B"},
			in: [][]string{
				{"a", "b"},
				{"c", "d"},
				{"e", "e\n"},
				{"f", "f\nf"},
				{"g", "g\ng\n"},
			},
			out: `+---+---+
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
		{
			title: []string{"A", "B"},
			in: [][]string{
				{"a", "a\t\t"},
				{"b", "\tb\t"},
				{"c", "\t\tc"},
			},
			out: `+---+-------------+
| A |      B      |
+---+-------------+
| a | a		  |
| b | 	b	  |
| c | 		c |
+---+-------------+
`,
		},
		{
			title: []string{"A", "B", "C"},
			in: [][]string{
				{"a", "云联 华北2（北京）\t\t", "ok"},
				{"b", "\t云联 印度尼西亚（雅加达）", "ok"},
				{"c", "云联 香港", "ok"},
			},
			out: `+---+------------------------------------+----+
| A |                 B                  | C  |
+---+------------------------------------+----+
| a | 云联 华北2（北京）		 | ok |
| b | 	云联 印度尼西亚（雅加达）        | ok |
| c | 云联 香港                          | ok |
+---+------------------------------------+----+
`,
		},
		{
			maxLineWidth: 24,
			title:        []string{"A", "B", "C"},
			in: [][]string{
				{"a", "云联 华北2（北京）\t\t", "ok"},
				{"b", "\t云联 印度尼西亚（雅加达）", "ok"},
				{"c", "云联 香港", "ok"},
			},
			out: `+---+-------------+----+
| A |      B      | C  |
+---+-------------+----+
| a | 云联 华北2  | ok |
|   | （北京）	  |    |
|   | 	          |    |
| b | 	云联 印度 | ok |
|   | 尼西亚（雅  |    |
|   | 加达）      |    |
| c | 云联 香港   | ok |
+---+-------------+----+
`,
		},
		{
			maxLineWidth: 16,
			title:        []string{"A", "B", "C"},
			in: [][]string{
				{"a", "tttttttttttttttttttttttttttt\t\t", "ok"},
				{"b", "\tttttttttttttttttt", "ok"},
				{"c", "ttttttttt", "ok"},
			},
			out: `+---+-----+----+
| A |  B  | C  |
+---+-----+----+
| a | ttt | ok |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | t	  |    |
|   | 	  |    |
| b | 	t | ok |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | ttt |    |
|   | t   |    |
| c | ttt | ok |
|   | ttt |    |
|   | ttt |    |
+---+-----+----+
`,
		},
		{
			maxLineWidth: 5,
			title:        []string{"A", "B", "C"},
			in: [][]string{
				{"a", "tt", "ok"},
				{"b", "\ttttt", "ok"},
				{"c", "t", "ok"},
			},
			out: `+---+---+---+
| A | B | C |
+---+---+---+
| a | t | o |
|   | t | k |
| b | 	 | o |
|   | t | k |
|   | t |   |
|   | t |   |
|   | t |   |
| c | t | o |
|   |   | k |
+---+---+---+
`,
		},
	}
	for _, c := range cases {
		pt := NewPrettyTable(c.title)
		if c.maxLineWidth > 0 {
			pt.MaxLineWidth(c.maxLineWidth)
		}
		out := pt.GetString(c.in)
		if out != c.out {
			t.Errorf("got != want\n%s\n(%d)\n !=\n%s\n(%d)\n", out, len(out), c.out, len(c.out))
		}
	}
}
