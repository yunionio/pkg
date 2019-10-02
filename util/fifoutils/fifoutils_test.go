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

package fifoutils

import (
	"reflect"
	"testing"
)

func TestFIFO_Pop(t *testing.T) {
	fifo := NewFIFO()
	fifo.Push(1)
	fifo.Push(2)
	fifo.Push(3)
	fifo.Push(4)
	ff := func(want interface{}, l int) {
		got := fifo.Pop()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("bad val: got %#v, want %#v", got, want)
		}
		if fifo.len != l {
			t.Fatalf("bad len: got %d, want %d", fifo.len, l)
		}
		for i := l; i < len(fifo.array); i++ {
			if fifo.array[i] != nil {
				t.Fatalf("bad val at %d, want nil, got %#v", i, fifo.array[i])
			}
		}
	}
	ff(1, 3)
	ff(2, 2)
	ff(3, 1)
	ff(4, 0)
	ff(nil, 0)
	ff(nil, 0)
}
