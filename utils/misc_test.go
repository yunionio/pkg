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

package utils

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

type testObj struct {
	id   string
	name string
}

func TestToDict(t *testing.T) {
	idSelect := func(obj interface{}) (string, error) {
		return obj.(*testObj).id, nil
	}

	idErr := func(obj interface{}) (string, error) {
		return obj.(*testObj).id, fmt.Errorf("Error")
	}

	type args struct {
		items []interface{}
		ks    selectFunc
	}

	obj1 := &testObj{"1", "name1"}
	obj2 := &testObj{"2", "name2"}
	items := []interface{}{obj1, obj2}
	args1 := args{items, idSelect}
	args2 := args{items, idErr}
	exp1 := map[string]interface{}{"1": obj1, "2": obj2}

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{"Basic", args1, exp1, false},
		{"WithError", args2, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDict(tt.args.items, tt.args.ks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToDict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroupBy(t *testing.T) {
	obj1 := &testObj{"1", "name1"}
	obj2 := &testObj{"2", "name1"}
	obj3 := &testObj{"3", "name2"}
	obj4 := &testObj{"4", "name2"}
	items := []interface{}{obj1, obj2, obj3, obj4}

	nameSelect := func(obj interface{}) (string, error) {
		return obj.(*testObj).name, nil
	}

	type args struct {
		items []interface{}
		ks    selectFunc
	}

	args1 := args{items, nameSelect}
	exp1 := map[string][]interface{}{"name1": {obj1, obj2}, "name2": {obj3, obj4}}

	tests := []struct {
		name    string
		args    args
		want    map[string][]interface{}
		wantErr bool
	}{
		{"Basic", args1, exp1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GroupBy(tt.args.items, tt.args.ks)
			if (err != nil) != tt.wantErr {
				t.Errorf("GroupBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupBy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectDistinct(t *testing.T) {
	obj1 := &testObj{"1", "name1"}
	obj2 := &testObj{"2", "name1"}
	obj3 := &testObj{"3", "name2"}
	obj4 := &testObj{"4", "name2"}
	items := []interface{}{obj1, obj2, obj3, obj4}

	nameSelect := func(obj interface{}) (string, error) {
		return obj.(*testObj).name, nil
	}

	type args struct {
		items []interface{}
		ks    selectFunc
	}

	args1 := args{items, nameSelect}
	exp1 := []string{"name1", "name2"}

	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr error
	}{
		{"Basic", args1, exp1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectDistinct(tt.args.items, tt.args.ks)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("SelectDistinct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SelectDistinct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubDict(t *testing.T) {
	obj1 := testObj{"1", "name1"}
	obj2 := testObj{"2", "name2"}
	obj3 := testObj{"3", "name3"}
	obj4 := testObj{"4", "name4"}
	items := map[string][]testObj{"1": {obj1, obj2}, "2": {obj2, obj3}, "3": {obj3, obj4}, "4": {obj4, obj1}}
	items2 := make(map[string][]interface{}, 0)
	for key, value := range items {
		interfaces := make([]interface{}, 0)
		for _, val := range value {
			interfaces = append(interfaces, val)
		}
		items2[key] = interfaces
	}
	nameSelect := func(obj interface{}) (string, error) {
		return obj.(*testObj).name, nil
	}

	type args struct {
		items map[string][]interface{}
		ks    selectFunc
	}

	args1 := args{items2, nameSelect}
	exp1 := map[string][]testObj{"2": {obj2, obj3}, "3": {obj3, obj4}}

	tests := []struct {
		name    string
		args    args
		want    map[string][]testObj
		wantErr bool
	}{
		{"Basic", args1, exp1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SubDict(tt.args.items, "2", "3")
			if (err != nil) != tt.wantErr {
				t.Errorf("SubDict() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("SubDict() = %v, want %v", got, tt.want)
			// }
			for key := range got {
				if _, ok := tt.want[key]; !ok {
					t.Errorf("SubDict() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

type testHasPrefix struct {
	stringValue string
	prefixValue string
	want        bool
}

func TestHasPrefix(t *testing.T) {
	tests := []testHasPrefix{
		{"meituan", "mei", true},
		{"meituan", "hei", false},
	}
	for _, tt := range tests {
		if ok := HasPrefix(tt.stringValue, tt.prefixValue); ok != tt.want {
			t.Errorf("HasPrefix() = %v, want %v", ok, tt.want)
		}
	}
}

type testHasSuffix struct {
	stringValue string
	suffixValue string
	want        bool
}

func TestHasSuffix(t *testing.T) {
	tests := []testHasSuffix{
		{"meituan", "tuan", true},
		{"meituan", "quan", false},
	}
	for _, tt := range tests {
		if ok := HasSuffix(tt.stringValue, tt.suffixValue); ok != tt.want {
			t.Errorf("HasSuffix(%v) = %v, want %v", tt, ok, tt.want)
		}
	}
}

type testIPMatch struct {
	ip   string
	want bool
}

func TestIsMatchIP4(t *testing.T) {
	tests := []testIPMatch{
		{"1.1.1.1", true},
		{"255.1.1.1", true},
		{"256.1.1.1", true},
		{"0.1.0.1", true},
		{"0.1.0.1.3", false},
	}
	for _, tt := range tests {
		if ok := IsMatchIP4(tt.ip); ok != tt.want {
			t.Errorf("IsMatchIP4(%s) = %v, want %v", tt.ip, ok, tt.want)
		}
	}
}

func TestIsMatchIP6(t *testing.T) {
	tests := []testIPMatch{
		{"1.1.1.1.1.1", false},
		{"255.1.1.1.1.1", false},
		{"256.1.1.1.1.1", false},
		{"fe80::7c16:3bff:fe33:a2d4", false},
		{"ff.1.1.1.ff.ff", false},
		{"fh.1.1.1.ff.ff", false},
		{"0.1.0.1", false},
		{"0.1.0.1.3.1.1.1.1", false},
	}
	for _, tt := range tests {
		if ok := IsMatchIP6(tt.ip); ok != tt.want {
			t.Errorf("IsMatchIP6(%s) = %v, want %v", tt.ip, ok, tt.want)
		}
	}
}

type testMacMatch struct {
	mac  string
	want bool
}

func TestIsMatchCompactMacAddr(t *testing.T) {
	tests := []testMacMatch{
		{"4E:17:A9:3C:14:4E", false},
		{"00:33:00:FF:F1:83", false},
		{"256.1.1.1.1.1", false},
		{"ff.1.1.1.ff.ff", false},
		{"fh.1.1.1.ff.ff", false},
		{"0.1.0.1", false},
		{"0.1.0.1.3.1.1.1.1", false},
	}
	for _, tt := range tests {
		if ok := IsMatchCompactMacAddr(tt.mac); ok != tt.want {
			t.Errorf("IsMatchCompactMacAddr(%s) = %v, want %v", tt.mac, ok, tt.want)
		}
	}
}

func TestIsMatchMacAddr(t *testing.T) {
	tests := []testMacMatch{
		{"4E:17:A9:3C:14:4E", true},
		{"00:33:00:FF:F1:83", true},
		{"256.1.1.1.1.1", false},
		{"ff.1.1.1.ff.ff", false},
		{"fh.1.1.1.ff.ff", false},
		{"0.1.0.1", false},
		{"0.1.0.1.3.1.1.1.1", false},
	}
	for _, tt := range tests {
		if ok := IsMatchMacAddr(tt.mac); ok != tt.want {
			t.Errorf("IsMatchMacAddr(%s) = %v, want %v", tt.mac, ok, tt.want)
		}
	}
}

type testNums struct {
	num1 int64
	num2 int64
	want int64
}

func TestMax(t *testing.T) {
	tests := []testNums{
		{12, 45, 45},
		{12, -45, 12},
		{0, -45, 0},
	}
	for _, tt := range tests {
		if max := Max(tt.num1, tt.num2); max != tt.want {
			t.Errorf("Max(%d %d) = %v, want %v", tt.num1, tt.num2, max, tt.want)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []testNums{
		{12, 45, 12},
		{12, -45, -45},
		{0, -45, -45},
	}
	for _, tt := range tests {
		if max := Min(tt.num1, tt.num2); max != tt.want {
			t.Errorf("Min(%d %d) = %v, want %v", tt.num1, tt.num2, max, tt.want)
		}
	}
}

type testExitAddress struct {
	ip   string
	want bool
}

func TestIsExitAddress(t *testing.T) {
	tests := []testExitAddress{
		{"127.0.0.1", false},
		{"192.168.8.9", false},
		{"172.16.0.2", false},
		{"172.32.255.255", false},
		{"169.254.0.1", false},
		{"169.254.255.255", false},
		{"172.32.255.255", false},
		{"8.8.8.8", true},
		{"8.8.8.8.8", false},
		{"0.1.0.1", true},
		{"0.1.0", false},
	}
	for _, tt := range tests {
		if ok := IsExitAddress(tt.ip); ok != tt.want {
			t.Errorf("IsExitAddress(%s) = %v, want %v", tt.ip, ok, tt.want)
		}
	}
}

func TestToInt64E(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "string",
			args:    args{"12"},
			want:    12,
			wantErr: false,
		},
		{
			name:    "stringWithError",
			args:    args{"ten"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToInt64E(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToInt64E() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToInt64E() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDurationE(t *testing.T) {
	type args struct {
		i interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantD   time.Duration
		wantErr bool
	}{
		{
			name:    "seconds",
			args:    args{"2s"},
			wantD:   2 * time.Second,
			wantErr: false,
		},
		{
			name:    "minutes",
			args:    args{"2m"},
			wantD:   2 * time.Minute,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, err := ToDurationE(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDurationE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotD != tt.wantD {
				t.Errorf("ToDurationE() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}

func TestIsMatchInteger(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match integer",
			args: args{"012"},
			want: true,
		},
		{
			name: "not match integer",
			args: args{"f012"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMatchInteger(tt.args.s); got != tt.want {
				t.Errorf("IsMatchInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMatchFloat(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match float",
			args: args{"012"},
			want: true,
		},
		{
			name: "match float",
			args: args{"012"},
			want: true,
		},
		{
			name: "not match float",
			args: args{"nofloat"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMatchFloat(tt.args.s); got != tt.want {
				t.Errorf("IsMatchFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSize(t *testing.T) {
	type args struct {
		sizeStr     string
		defaultSize string
		base        int64
	}
	tests := []struct {
		name     string
		args     args
		wantSize int64
		wantErr  bool
	}{
		{
			name:     "parse 1024, defaultSize b",
			args:     args{"1024", "b", 1024},
			wantSize: 1024,
			wantErr:  false,
		},
		{
			name:     "parse 1024B, defaultSize K",
			args:     args{"1024B", "K", 1024},
			wantSize: 1024,
			wantErr:  false,
		},
		{
			name:     "parse 1024K, defaultSize K",
			args:     args{"1024K", "K", 1024},
			wantSize: 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "parse 1024G, defaultSize K",
			args:     args{"1024G", "K", 1024},
			wantSize: 1024 * 1024 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "test err",
			args:     args{"G1024G", "K", 1024},
			wantSize: 0,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, err := GetSize(tt.args.sizeStr, tt.args.defaultSize, tt.args.base)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSize != tt.wantSize {
				t.Errorf("GetSize() = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func TestGetBytes(t *testing.T) {
	type args struct {
		sizeStr string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "parse 1024 with error",
			args:    args{"1024"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "parse 1024g",
			args:    args{"1024g"},
			want:    1024 * 1024 * 1024 * 1024,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBytes(tt.args.sizeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSizeMB(t *testing.T) {
	type args struct {
		sizeStr     string
		defaultSize string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "get size 100, unit M",
			args:    args{"100", "M"},
			want:    100,
			wantErr: false,
		},
		{
			name:    "get size 10M",
			args:    args{"10M", "M"},
			want:    10,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSizeMB(tt.args.sizeStr, tt.args.defaultSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSizeMB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSizeMB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransSQLAchemyURL(t *testing.T) {
	type args struct {
		pySQLSrc string
	}
	tests := []struct {
		name    string
		args    args
		wantRet string
		wantErr bool
	}{
		{
			name:    "should convert",
			args:    args{"mysql+pymysql://root:root@127.0.0.1:3306/mclouds?charset=utf8"},
			wantRet: "root:root@tcp(127.0.0.1:3306)/mclouds?charset=utf8&parseTime=True",
			wantErr: false,
		},
		{
			name:    "bare",
			args:    args{"root:root@127.0.0.1:3306/mclouds?charset=utf8"},
			wantRet: "root:root@127.0.0.1:3306/mclouds?charset=utf8",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotRet, err := TransSQLAchemyURL(tt.args.pySQLSrc)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransSQLAchemyURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRet != tt.wantRet {
				t.Errorf("TransSQLAchemyURL() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func TestComposeURL(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty paths",
			args: args{[]string{}},
			want: "",
		},
		{
			name: "one empty path",
			args: args{[]string{""}},
			want: "",
		},
		{
			name: "multiple paths",
			args: args{[]string{"specs", "model", "ident", "action"}},
			want: "/specs/model/ident/action",
		},
		{
			name: "multiple paths with empty string",
			args: args{[]string{"specs", "model", "", "", "action"}},
			want: "/specs/model",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeURL(tt.args.paths...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComposeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSizeGB(t *testing.T) {
	type args struct {
		sizeStr     string
		defaultSize string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "get size 1024, unit M",
			args:    args{"1024", "M"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "get size 10240M",
			args:    args{"10240M", "M"},
			want:    10,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSizeGB(tt.args.sizeStr, tt.args.defaultSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSizeGB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSizeGB() = %v, want %v", got, tt.want)
			}
		})
	}
}
