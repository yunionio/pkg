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

package secrules

import (
	"testing"
)

func TestPorts(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		ps := newPortsFromInts(1, 3)
		if !ps.contains(1) {
			t.Fatalf("contains")
		}
	})
	t.Run("dedup", func(t *testing.T) {
		ps := newPortsFromInts(1, 3, 3, 4, 1, 3, 5)
		ps1 := ps.dedup()
		if len(ps1) != 4 || !ps.containsPorts(newPortsFromInts(1, 3, 4, 5)) {
			t.Fatalf("dedup")
		}
	})
	t.Run("subs ports", func(t *testing.T) {
		ps0 := newPortsFromInts(1, 3, 4)
		ps1 := newPortsFromInts(1, 4, 9)
		left, sub := ps0.substractPorts(ps1)
		if len(left) != 1 || !left.contains(3) {
			t.Fatalf("bad `left`")
		}
		if len(sub) != 2 || !sub.contains(1) || !sub.contains(4) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("subs range", func(t *testing.T) {
		ps := newPortsFromInts(4, 7, 8)
		t.Run("disjoint (left)", func(t *testing.T) {
			pr := newPortRange(9, 10)
			left, sub := ps.substractPortRange(pr)
			if len(left) != 3 || !left.sameAs(ps) {
				t.Fatalf("bad `left`")
			}
			if len(sub) != 0 {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("overlap (left)", func(t *testing.T) {
			pr := newPortRange(1, 4)
			left, sub := ps.substractPortRange(pr)
			if len(left) != 2 || !left.sameAs(newPortsFromInts(7, 8)) {
				t.Fatalf("bad `left`")
			}
			if len(sub) != 1 || !sub.contains(4) {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("contained by", func(t *testing.T) {
			pr := newPortRange(1, 10)
			left, sub := ps.substractPortRange(pr)
			if len(left) != 0 {
				t.Fatalf("bad `left`")
			}
			if len(sub) != len(ps) || !sub.sameAs(ps) {
				t.Fatalf("bad `sub`")
			}
		})
	})
}

func TestPortRange(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		pr := newPortRange(5, 7)
		if !pr.contains(5) {
			t.Fatalf("contains")
		}
		if pr.count() != 3 {
			t.Fatalf("count")
		}
		if !pr.containsRange(newPortRange(5, 6)) {
			t.Fatalf("containsRange")
		}
		if pr.containsRange(newPortRange(4, 6)) {
			t.Fatalf("!containsRange")
		}
	})
	t.Run("subs ports", func(t *testing.T) {
		pr := newPortRange(5, 7)
		ps := newPortsFromInts(1, 6)
		lefts, sub := pr.substractPorts(ps)
		if len(lefts) != 2 || !lefts[0].equals(newPortRange(5, 5)) || !lefts[1].equals(newPortRange(7, 7)) {
			t.Fatalf("bad `lefts`")
		}
		if len(sub) != 1 || !sub.contains(6) {
			t.Fatalf("bad `sub`")
		}
	})
	t.Run("subs range", func(t *testing.T) {
		pr := newPortRange(5, 7)
		t.Run("disjoint (left)", func(t *testing.T) {
			pr1 := newPortRange(2, 4)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 1 || !lefts[0].equals(pr) {
				t.Fatalf("bad `lefts`")
			}
			if sub != nil && sub.count() != 0 {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("overlap (cut right)", func(t *testing.T) {
			pr1 := newPortRange(2, 5)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 1 || !lefts[0].equals(newPortRange(6, 7)) {
				t.Fatalf("bad `lefts`")
			}
			if sub == nil || !sub.equals(newPortRange(5, 5)) {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("contains (true subset)", func(t *testing.T) {
			pr1 := newPortRange(6, 6)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 2 || !lefts[0].equals(newPortRange(5, 5)) || !lefts[1].equals(newPortRange(7, 7)) {
				t.Fatalf("bad `lefts`")
			}
			if sub == nil || !sub.equals(newPortRange(6, 6)) {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("contains (align left)", func(t *testing.T) {
			pr1 := newPortRange(5, 6)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 1 || !lefts[0].equals(newPortRange(7, 7)) {
				t.Fatalf("bad `lefts`")
			}
			if sub == nil || !sub.equals(newPortRange(5, 6)) {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("contains (align right)", func(t *testing.T) {
			pr1 := newPortRange(6, 7)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 1 || !lefts[0].equals(newPortRange(5, 5)) {
				t.Fatalf("bad `lefts`")
			}
			if sub == nil || !sub.equals(newPortRange(6, 7)) {
				t.Fatalf("bad `sub`")
			}
		})
		t.Run("contained by", func(t *testing.T) {
			pr1 := newPortRange(2, 9)
			lefts, sub := pr.substractPortRange(pr1)
			if len(lefts) != 0 {
				t.Fatalf("bad `lefts`")
			}
			if sub == nil || !sub.equals(pr) {
				t.Fatalf("bad `sub`")
			}
		})
	})
}
