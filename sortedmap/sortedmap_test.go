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

package sortedmap

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSortedMap(t *testing.T) {
	ss := NewSortedMapFromMap(map[string]interface{}{
		"Go":     "Go",
		"Bravo":  "Bravo",
		"Gopher": "Gopher",
		"Alpha":  "Alpha",
		"Grin":   "Grin",
		"Delta":  "Delta",
	})
	// Alpha Bravo Delta Go Gopher Grin
	// 0     1     2     3  4      5
	cases := []struct {
		Needle string
		Index  int
		Want   bool
	}{
		{"Go", 3, true},
		{"Bravo", 1, true},
		{"Gopher", 4, true},
		{"Alpha", 0, true},
		{"Grin", 5, true},
		{"Delta", 2, true},
		{"Go1", 4, false},
		{"G", 3, false},
		{"A", 0, false},
		{"T", 6, false},
	}
	for _, c := range cases {
		idx, find := ss.find(c.Needle)
		if idx != c.Index || find != c.Want {
			t.Errorf("%s: want: %d %v got: %d %v", c.Needle, c.Index, c.Want, idx, find)
		}
	}
}

func TestSplitMaps(t *testing.T) {
	cases := []struct {
		input  map[string]interface{}
		input2 map[string]interface{}
		want1  []string
		want2  []string
		want3  []string
	}{
		{
			input:  map[string]interface{}{"Go": 1, "Bravo": 1, "Gopher": 1, "Alpha": 1, "Grin": 1, "Delta": 1},
			input2: map[string]interface{}{"Go2": 2, "Bravo": 2, "Gopher": 2, "Alpha1": 2, "Grin": 2, "Delt": 2},
			want1:  []string{"Alpha", "Delta", "Go"},
			want2:  []string{"Bravo", "Gopher", "Grin"},
			want3:  []string{"Alpha1", "Delt", "Go2"},
		},
	}

	for _, c := range cases {
		ss1 := NewSortedMapFromMap(c.input)
		ss2 := NewSortedMapFromMap(c.input2)

		a, b1, b2, d := Split(ss1, ss2)
		if !reflect.DeepEqual(a.Keys(), c.want1) {
			t.Fatalf("A-B got %s want %s", a.Keys(), c.want1)
		}
		if !reflect.DeepEqual(b1.Keys(), c.want2) {
			t.Fatalf("A and B in A got %s want %s", b1.Keys(), c.want2)
		}
		if !reflect.DeepEqual(b2.Keys(), c.want2) {
			t.Fatalf("A and B in B got %s want %s", b2.Keys(), c.want2)
		}
		if !reflect.DeepEqual(d.Keys(), c.want3) {
			t.Fatalf("B-A got %s want %s", d.Keys(), c.want3)
		}
	}
}

func TestMergeMaps(t *testing.T) {
	cases := []struct {
		input  map[string]interface{}
		input2 map[string]interface{}
		want   []string
	}{
		{
			input:  map[string]interface{}{"Go": 1, "Bravo": 1, "Gopher": 1, "Alpha": 1, "Grin": 1, "Delta": 1},
			input2: map[string]interface{}{"Go2": 2, "Bravo": 2, "Gopher": 2, "Alpha1": 2, "Grin": 2, "Delt": 2},
			want:   []string{"Alpha", "Alpha1", "Bravo", "Delt", "Delta", "Go", "Go2", "Gopher", "Grin"},
		},
	}
	for _, c := range cases {
		ss1 := NewSortedMapFromMap(c.input)
		ss2 := NewSortedMapFromMap(c.input2)

		m := Merge(ss1, ss2)
		if !reflect.DeepEqual(m.Keys(), c.want) {
			t.Fatalf("merge got %s want %s", m.Keys(), c.want)
		}
	}
}

func TestSortedStringsAppend(t *testing.T) {
	cases := []struct {
		in   []string
		ele  []string
		want []string
	}{
		{
			in:   []string{"Alpha", "Bravo", "Go"},
			ele:  []string{"Go2"},
			want: []string{"Alpha", "Bravo", "Go", "Go2"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go2"},
			ele:  []string{"Go"},
			want: []string{"Alpha", "Bravo", "Go", "Go2"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go2"},
			ele:  []string{"Aaaa", "Go"},
			want: []string{"Aaaa", "Alpha", "Bravo", "Go", "Go2"},
		},
	}
	for _, c := range cases {
		ss := NewSortedMap()
		for _, v := range c.in {
			ss = Add(ss, v, 1)
		}
		for _, v := range c.ele {
			ss = Add(ss, v, 1)
		}
		got := ss.Keys()
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("want: %s got: %s", c.want, got)
		}
	}
}

func TestSortedStringsRemove(t *testing.T) {
	cases := []struct {
		in   []string
		ele  []string
		want []string
	}{
		{
			in:   []string{"Alpha", "Bravo", "Go"},
			ele:  []string{"Go", "Go2"},
			want: []string{"Alpha", "Bravo"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go2"},
			ele:  []string{"Go"},
			want: []string{"Alpha", "Bravo", "Go2"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go", "Go2"},
			ele:  []string{"Aaaa", "Alpha"},
			want: []string{"Bravo", "Go", "Go2"},
		},
	}
	for _, c := range cases {
		ss := NewSortedMap()
		for _, v := range c.in {
			ss = Add(ss, v, 1)
		}
		for _, v := range c.ele {
			ss, _ = Delete(ss, v)
		}
		got := ss.Keys()
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("want: %s got: %s", c.want, got)
		}
	}
}

func TestSortedStringsDeleteIgnoreCase(t *testing.T) {
	cases := []struct {
		in   []string
		ele  []string
		want []string
	}{
		{
			in:   []string{"Alpha", "Bravo", "Go"},
			ele:  []string{"go", "Go2"},
			want: []string{"Alpha", "Bravo"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go2"},
			ele:  []string{"Go"},
			want: []string{"Alpha", "Bravo", "Go2"},
		},
		{
			in:   []string{"Alpha", "Bravo", "Go", "Go2"},
			ele:  []string{"Aaaa", "alpha"},
			want: []string{"Bravo", "Go", "Go2"},
		},
	}
	for _, c := range cases {
		ss := NewSortedMap()
		for _, v := range c.in {
			ss = Add(ss, v, 1)
		}
		for _, v := range c.ele {
			ss, _, _ = DeleteIgnoreCase(ss, v)
		}
		got := ss.Keys()
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("want: %s got: %s", c.want, got)
		}
	}
}

func TestCopy(t *testing.T) {
	const (
		alphabet  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		alphabet1 = "AABCDEFGHIJKLMNOPQRSTUVWXY"
		alphabet2 = "BCDEFGHIJKLMNOPQRSTUVWXYZZ"
	)
	s := []byte(alphabet)
	copy(s[1:], s)
	if string(s) != alphabet1 {
		t.Fatalf("right shift fail")
	}
	w := []byte(alphabet)
	copy(w, w[1:])
	if string(w) != alphabet2 {
		t.Fatalf("left shift fail")
	}
}

func TestIterator(t *testing.T) {
	input := map[string]interface{}{"Go": 1, "Bravo": 1, "Gopher": 1, "Alpha": 1, "Grin": 1, "Delta": 1}
	ss := NewSortedMapFromMap(input)
	keys := make([]string, 0)
	for iter := NewIterator(ss); iter.HasMore(); iter.Next() {
		k, _ := iter.Get()
		keys = append(keys, k)
	}
	if !reflect.DeepEqual(keys, ss.Keys()) {
		t.Fatalf("keys not equal %s != %s", keys, ss.Keys())
	}
}

func BenchmarkAdd(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		input := make(map[string]interface{})
		keys := []string{
			"Go", "Bravo", "Gopher", "Alpha", "Grin", "Delta", "Delta2", "Alpha3",
		}
		for i := 0; i < size; i++ {
			input[fmt.Sprintf("%s-%d", keys[i%len(keys)], i)] = 1
		}
		ss := NewSortedMapFromMap(input)
		b.Run(fmt.Sprintf("sortedMap-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss = Add(ss, "Go2", 1)
				ss = Add(ss, "Aloha2", 1)
				ss = Add(ss, "Zero2", 1)
			}
		})
		b.Run(fmt.Sprintf("rawMap-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				input["Go2"] = 1
				input["Aloha2"] = 1
				input["Zero2"] = 1
			}
		})
	}
}

func BenchmarkDelete(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		keys := []string{
			"Go", "Bravo", "Bravo2", "Gopher", "Alpha", "Grin", "Delta", "Delt", "Delta2", "Alpha3", "Zoom",
		}
		input := make(map[string]interface{})
		for i := 0; i < size; i++ {
			input[fmt.Sprintf("%s-%d", keys[i%len(keys)], i)] = 1
		}
		ss := NewSortedMapFromMap(input)
		b.Run(fmt.Sprintf("sortedMap-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ss, _ = Delete(ss, "Grin")
				ss, _ = Delete(ss, "Bravo")
				ss, _ = Delete(ss, "Zoom")
			}
		})
		b.Run(fmt.Sprintf("rawMap-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				delete(input, "Grin")
				delete(input, "Bravo")
				delete(input, "Zoom")
			}
		})
	}
}
