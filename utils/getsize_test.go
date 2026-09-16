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

import "testing"

// An empty size string, with or without a default unit, is not a size.
func TestGetSizeRejectsEmpty(t *testing.T) {
	cases := []struct {
		name        string
		sizeStr     string
		defaultSize string
	}{
		{"both empty", "", ""},
		{"empty with default unit", "", "g"},
		{"empty default, non empty", "x", ""},
	}
	for _, c := range cases {
		size, err := GetSize(c.sizeStr, c.defaultSize, 1024)
		if err == nil {
			t.Errorf("%s: GetSize(%q, %q) returned no error, got %d", c.name, c.sizeStr, c.defaultSize, size)
		}
		if size != 0 {
			t.Errorf("%s: GetSize(%q, %q) returned %d together with an error", c.name, c.sizeStr, c.defaultSize, size)
		}
	}
}

func TestGetBytesRejectsEmpty(t *testing.T) {
	if size, err := GetBytes(""); err == nil {
		t.Errorf("GetBytes(\"\") returned no error, got %d", size)
	}
}

// A value that does not fit in an int64 must be reported, not wrapped around.
func TestGetSizeRejectsOverflow(t *testing.T) {
	cases := []struct {
		sizeStr     string
		defaultSize string
	}{
		{"17179869184g", "g"}, // 2^64
		{"9223372036854775807g", "g"},
		{"9223372036854775807t", "t"},
		{"99999999999999999999g", "g"},
	}
	for _, c := range cases {
		size, err := GetSize(c.sizeStr, c.defaultSize, 1024)
		if err == nil {
			t.Errorf("GetSize(%q, %q) returned no error, got %d", c.sizeStr, c.defaultSize, size)
		}
	}
}

// Values that fit must be returned exactly as before.
func TestGetSizeAcceptedValues(t *testing.T) {
	cases := []struct {
		sizeStr     string
		defaultSize string
		want        int64
	}{
		{"1024", "b", 1024},
		{"1024B", "", 1024},
		{"1024k", "", 1024 * 1024},
		{"1024K", "", 1024 * 1024},
		{"1m", "", 1024 * 1024},
		{"1g", "", 1024 * 1024 * 1024},
		{"1024G", "", 1024 * 1024 * 1024 * 1024},
		{"1t", "", 1024 * 1024 * 1024 * 1024},
		{"0g", "", 0},
	}
	for _, c := range cases {
		got, err := GetSize(c.sizeStr, c.defaultSize, 1024)
		if err != nil {
			t.Errorf("GetSize(%q, %q): %v", c.sizeStr, c.defaultSize, err)
			continue
		}
		if got != c.want {
			t.Errorf("GetSize(%q, %q) = %d, want %d", c.sizeStr, c.defaultSize, got, c.want)
		}
	}
}

// An unknown unit is still reported.
func TestGetSizeRejectsUnknownUnit(t *testing.T) {
	if size, err := GetSize("1024z", "", 1024); err == nil {
		t.Errorf("GetSize(\"1024z\") returned no error, got %d", size)
	}
}
