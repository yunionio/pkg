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

package gotypes

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"yunion.io/x/pkg/util/timeutils"
)

func TestParseValue(t *testing.T) {
	cases := []struct {
		in  string
		tp  reflect.Type
		out reflect.Value
	}{
		{"true", BoolType, reflect.ValueOf(true)},
		{"100", Int64Type, reflect.ValueOf(int64(100))},
		{"100", Int8Type, reflect.ValueOf(int8(100))},
		{"100", Uint8Type, reflect.ValueOf(uint8(100))},
		{"string", StringType, reflect.ValueOf("string")},
		{"3.14", Float32Type, reflect.ValueOf(float32(3.14))},
		{"3.14", Float64Type, reflect.ValueOf(float64(3.14))},
	}
	for _, c := range cases {
		get, e := ParseValue(c.in, c.tp)
		if e != nil {
			t.Errorf("ParseValue error %s", e)
		}
		if get.Interface() != c.out.Interface() {
			t.Errorf("ParseValue %s %s = %v, not %v", c.in, c.tp, get, c.out)
		}
	}
}

func TestSetValueInt(t *testing.T) {
	cases := []struct {
		In  int32
		Val string
		Out int32
	}{
		{1, "30", 30},
		{1, "200", 200},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := SetValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if c.In != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", c.In, c.Val, c.Out)
		}
	}
}

func TestSetValueFloat(t *testing.T) {
	cases := []struct {
		In  float32
		Val string
		Out float32
	}{
		{1, "30.0", 30.0},
		{1, "200.0", 200.0},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := SetValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if c.In != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", c.In, c.Val, c.Out)
		}
	}
}

func TestSetValueString(t *testing.T) {
	cases := []struct {
		In  string
		Val string
		Out string
	}{
		{"a", "bcd", "bcd"},
		{"h", "hello", "hello"},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := SetValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if c.In != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", c.In, c.Val, c.Out)
		}
	}
}

func TestSetValueStringArray(t *testing.T) {
	cases := []struct {
		In  []string
		Val string
		Out []string
	}{
		{[]string{}, "bcd", []string{"bcd"}},
		{[]string{"a"}, "hello", []string{"a", "hello"}},
		{[]string{"a"}, "hello,world", []string{"a", "hello", "world"}},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := SetValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if !reflect.DeepEqual(c.In, c.Out) {
			t.Errorf("SetValue fail %v (%s) != %v", c.In, c.Val, c.Out)
		}
	}
}

func TestSetValueIntArray(t *testing.T) {
	cases := []struct {
		In  []int
		Val string
		Out []int
	}{
		{[]int{}, "1,2", []int{1, 2}},
		{[]int{1}, "2", []int{1, 2}},
		{[]int{1}, "2,3", []int{1, 2, 3}},
		{[]int{1, 2}, "4,5", []int{1, 2, 4, 5}},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := SetValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if !reflect.DeepEqual(c.In, c.Out) {
			t.Errorf("SetValue fail %v (%s) != %v", c.In, c.Val, c.Out)
		}
	}
}

func TestAppendValueInt(t *testing.T) {
	cases := []struct {
		In  []int
		Val string
		Out int
	}{
		{make([]int, 0), "3", 3},
		{make([]int, 0), "100", 100},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := AppendValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if len(c.In) == 0 || c.In[0] != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", ref1, c.Val, c.Out)
		}
	}
}

func TestAppendValueFloat(t *testing.T) {
	cases := []struct {
		In  []float32
		Val string
		Out float32
	}{
		{make([]float32, 0), "3.0", 3.0},
		{make([]float32, 0), "100.0", 100.0},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := AppendValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if len(c.In) == 0 || c.In[0] != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", ref1, c.Val, c.Out)
		}
	}
}

func TestAppendValueString(t *testing.T) {
	cases := []struct {
		In  []string
		Val string
		Out string
	}{
		{make([]string, 0), "3.0", "3.0"},
		{make([]string, 0), "100.0", "100.0"},
	}
	for _, c := range cases {
		ref := reflect.ValueOf(&c).Elem()
		ref1 := ref.Field(0)
		e := AppendValue(ref1, c.Val)
		if e != nil {
			t.Errorf("SetValue error %s", e)
		}
		if len(c.In) == 0 || c.In[0] != c.Out {
			t.Errorf("SetValue fail %v (%s) != %v", ref1, c.Val, c.Out)
		}
	}
}

func TestInCollection(t *testing.T) {
	cases := []struct {
		obj    string
		arr    []string
		result bool
	}{
		{"abc", []string{"abc", "bcd"}, true},
		{"abc", []string{"a1bc", "bdc"}, false},
		{"", []string{" "}, false},
		{"", []string{}, false},
	}
	for _, c := range cases {
		if InCollection(c.obj, c.arr) != c.result {
			t.Errorf("%s in %s != %v", c.obj, c.arr, c.result)
		}
	}
}

func TestGetInstanceTypeName(t *testing.T) {
	{
		var (
			a    int32
			want = "int32"
		)
		if got := GetInstanceTypeName(a); got != want {
			t.Fatalf("want %s, got %s", want, got)
		}
	}

	{
		type STestStruct struct{}
		var (
			want = "STestStruct"
			a    = STestStruct{}
		)
		if got := GetInstanceTypeName(a); got != want {
			t.Fatalf("want %s, got %s", want, got)
		}
		if got := GetInstanceTypeName(&a); got != want {
			t.Fatalf("ptr type name: want %s, got %s", want, got)
		}
	}
}

func TestIsNil(t *testing.T) {
	var nilInterface interface{}
	var nilMap map[int]int
	var nilSlice []int
	var nilFunc func()
	var nilChannel chan int
	var nilPtr *int
	nils := []struct {
		val       interface{}
		equalsNil bool
		isNil     bool
	}{
		{val: nil, equalsNil: true, isNil: true},
		{val: nilInterface, equalsNil: true, isNil: true},
		{val: nilMap, equalsNil: false, isNil: true},
		{val: nilSlice, equalsNil: false, isNil: true},
		{val: nilFunc, equalsNil: false, isNil: true},
		{val: nilChannel, equalsNil: false, isNil: true},
		{val: nilPtr, equalsNil: false, isNil: true},
	}
	for _, n := range nils {
		var got bool
		got = n.val == nil
		if got != n.equalsNil {
			t.Errorf("equalsNil: want %v, got %v", n.equalsNil, got)
		}
		got = IsNil(n.val)
		if got != n.isNil {
			t.Errorf("isNil: want %v, got %v", n.isNil, got)
		}
	}
	cases := []struct {
		val   interface{}
		isNil bool
	}{
		{
			val:   [0]int{},
			isNil: false,
		},
	}
	for _, c := range cases {
		got := IsNil(c.val)
		if got != c.isNil {
			t.Errorf("want %v, got %v", c.isNil, got)
		}
	}
}

func TestParseSetValuePtr(t *testing.T) {
	v := &struct {
		BoolPtr      *bool
		IntPtr       *int
		UintPtr      *uint
		Float32Ptr   *float32
		Float64Ptr   *float64
		StringPtr    *string
		TimeISOPtr   *time.Time
		TimeMySQLPtr *time.Time
	}{}
	// make them addressable
	rv := reflect.ValueOf(v).Elem()

	mustParseTimeStr := func(s string) time.Time {
		t, err := timeutils.ParseTimeStr(s)
		if err != nil {
			panic("parseTimeStr " + s + ": " + err.Error())
		}
		return t
	}
	cases := []struct {
		name   string
		valStr string
		want   interface{}
	}{
		{"BoolPtr", "true", true},
		{"BoolPtr", "false", false},
		{"IntPtr", "-100", int(-100)},
		{"UintPtr", "100", uint(100)},
		{"Float32Ptr", "100.1", float32(100.1)},
		{"Float64Ptr", "-100.1", float64(-100.1)},
		{"StringPtr", "holy", "holy"},
		{"TimeISOPtr", "2018-08-27T04:20:26Z", mustParseTimeStr("2018-08-27T04:20:26Z")},
		{"TimeMySQLPtr", "2018-08-27 12:20:26", mustParseTimeStr("2018-08-27 12:20:26")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rvv := rv.FieldByName(c.name)
			t.Run("SetValue", func(t *testing.T) {
				err := SetValue(rvv, c.valStr)
				if err != nil {
					t.Fatalf("SetValue(%q, %q): failed: %s", c.name, c.valStr, err)
				}
				got := rvv.Interface()
				if reflect.DeepEqual(got, c.want) {
					t.Fatalf("SetValue(%q, %q): failed: got %#v, want %#v",
						c.name, c.valStr, got, c.want)
				}
				got = rvv.Elem().Interface()
				if !reflect.DeepEqual(got, c.want) {
					t.Fatalf("SetValue(%q, %q): failed: got %#v, want %#v",
						c.name, c.valStr, got, c.want)
				}
			})
			t.Run("ParseValue", func(t *testing.T) {
				rvvParse, err := ParseValue(c.valStr, rvv.Type())
				if err != nil {
					t.Fatalf("ParseValue(%q, %q): %s", c.valStr, rvv.Type(), err)
				}
				want := rvv.Interface()
				if got := rvvParse.Interface(); !reflect.DeepEqual(got, want) {
					t.Fatalf("ParseValue(%q, %q): failed: got %#v, want %#v",
						c.valStr, rvv.Type(), got, want)
				}
			})
		})
	}
}

func TestConvertSliceElemType(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cases := []struct {
			name   string
			inArg0 interface{}
			inArg1 interface{}
			r      interface{}
		}{
			{
				name:   "interface{} to int",
				inArg0: []interface{}{int(1), int(2)},
				r:      []int{1, 2},
			},
			{
				name:   "int to interface{}",
				inArg0: []int{1, 2},
				inArg1: (*interface{})(nil),
				r:      []interface{}{int(1), int(2)},
			},
			{
				name:   "interface{} to int (array)",
				inArg0: [2]interface{}{int(1), int(2)},
				r:      []int{1, 2},
			},
			{
				name:   "int to interface{} (array)",
				inArg0: [2]int{1, 2},
				inArg1: (*interface{})(nil),
				r:      []interface{}{int(1), int(2)},
			},
			{
				name:   "interface{} to int (empty)",
				inArg0: []interface{}{},
				inArg1: (*int)(nil),
				r:      []int{},
			},
			{
				name:   "interface{} to int (empty, no arg1)",
				inArg0: []interface{}{},
				r:      []interface{}{},
			},
			{
				name:   "int to interface{} (empty)",
				inArg0: []int{},
				inArg1: (*interface{})(nil),
				r:      []interface{}{},
			},
			{
				name:   "int to interface{} (empty, no arg1)",
				inArg0: []int{},
				r:      []int{},
			},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got := ConvertSliceElemType(c.inArg0, c.inArg1)
				if !reflect.DeepEqual(got, c.r) {
					t.Errorf("want %#v, got %#v", c.r, got)
				}
			})
		}
	})
	t.Run("panic", func(t *testing.T) {
		cases := []struct {
			name   string
			inArg0 interface{}
			inArg1 interface{}
			substr string
		}{
			{
				name:   "arg0 not slice or array",
				inArg0: "",
				substr: "slice or array",
			},
			{
				name:   "arg1 not ptr kind",
				inArg0: []int{},
				inArg1: int(0),
				substr: "be ptr kind",
			},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				defer func() {
					if val := recover(); val == nil {
						t.Errorf("expecting panic, got nil")
					} else if s, ok := val.(string); !ok {
						t.Errorf("expecting panic with string, got %#v", val)
					} else if !strings.Contains(s, c.substr) {
						t.Errorf("want %q in panic string %q", c.substr, s)
					}
				}()
				ConvertSliceElemType(c.inArg0, c.inArg1)
			})
		}
	})
}
