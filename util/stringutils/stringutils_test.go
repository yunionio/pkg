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

package stringutils

import (
	"fmt"
	"testing"
)

func TestStringUtils(t *testing.T) {
	t.Logf("%s", UUID4())
	if Interface2String(nil) != "" {
		t.Errorf("Interface2String(nil) should be empty")
	}
	if Interface2String(2) != "2" {
		t.Errorf(`Interface2String(2) should be "2"`)
	}
	if Interface2String("test string") != "test string" {
		t.Errorf(`Interface2String("test string") should be "test string"`)
	}
	if Interface2String(fmt.Errorf("test error")) != "test error" {
		t.Errorf(`Interface2String(fmt.Errorf("test error")) should be "test error"`)
	}
	type TestStruct struct {
		Name   string
		Age    int
		Gender string
	}
	t.Logf("%s", Interface2String(TestStruct{Name: "micheal", Age: 24, Gender: "Male"}))
}

func TestParseNamePattern(t *testing.T) {
	cases := []struct {
		name       string
		match      string
		pattern    string
		patternLen int
	}{
		{
			name:       "guest",
			match:      "guest-%",
			pattern:    "guest-%d",
			patternLen: 0,
		},
		{
			name:       "guest##",
			match:      "guest%",
			pattern:    "guest%02d",
			patternLen: 2,
		},
		{
			name:       "guest##suf",
			match:      "guest%suf",
			pattern:    "guest%02dsuf",
			patternLen: 2,
		},
		{
			name:       "test-###",
			match:      "test-%",
			pattern:    "test-%03d",
			patternLen: 3,
		},
	}
	for _, c := range cases {
		match, pattern, patternLen := ParseNamePattern(c.name)
		if match != c.match {
			t.Errorf("match: want %s, got %s", c.match, match)
		}
		if pattern != c.pattern {
			t.Errorf("pattern: want %s, got %s", c.pattern, pattern)
		}
		if patternLen != c.patternLen {
			t.Errorf("patternLen: want %d, got %d", c.patternLen, patternLen)
		}
	}
}

func TestSplitKeyValueBySep(t *testing.T) {
	type args struct {
		line string
		sep  string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name:  "emptyInput",
			args:  args{line: "", sep: ":"},
			want:  "",
			want1: "",
		},
		{
			name:  "normalInput",
			args:  args{line: "key 1: value 2", sep: ":"},
			want:  "key 1",
			want1: "value 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := SplitKeyValueBySep(tt.args.line, tt.args.sep)
			if got != tt.want {
				t.Errorf("SplitKeyValueBySep() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("SplitKeyValueBySep() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestContainsWord(t *testing.T) {
	cases := []struct {
		str  string
		word string
		want bool
	}{
		{
			str:  "account_id",
			word: "account",
			want: false,
		},
		{
			str:  "account_id",
			word: "account_id",
			want: true,
		},
		{
			str:  "(toYear(usage_start_time)*100)+toMonth(usage_start_time)",
			word: "usage_start_time",
			want: true,
		},
		{
			str:  "(toYear(usage_start_time)*100)+toMonth(usage_start_time)",
			word: "usage_start",
			want: false,
		},
		{
			str:  "(toYear(usage_start_time)*100)+toMonth(usage_start_time)",
			word: "(usage_start_time)",
			want: true,
		},
	}
	for _, c := range cases {
		got := ContainsWord(c.str, c.word)
		if got != c.want {
			t.Errorf("str %s contains word %s got %v want %v", c.str, c.word, got, c.want)
		}
	}
}

func TestByte2Str(t *testing.T) {
	cases := []struct {
		in []byte
		want string
	}{
		{
			in: []byte{0x00},
			want: "00",
		},
		{
			in: []byte{0x00, 0xff},
			want: "00ff",
		},
		{
			in: []byte{0xfe, 0x80, 0x00, 0x00},
			want: "fe800000",
		},
	}
	for _, c := range cases {
		got := Bytes2Str(c.in)
		if got != c.want {
			t.Errorf("got %s want %s", got, c.want)
		}
	}
}
