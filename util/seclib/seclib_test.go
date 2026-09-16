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

package seclib

import (
	"strings"
	"testing"
)

func TestRandomPassword(t *testing.T) {
	t.Logf("%s", RandomPassword(12))
}

func TestRandomPassword2(t *testing.T) {
	t.Logf("%s", RandomPassword2(12))
}

// The generated value has to meet the complexity rules the function filters
// on, and the first character must come from the leading-character set.
func TestRandomPassword2Shape(t *testing.T) {
	for i := 0; i < 200; i++ {
		pw := RandomPassword2(12)
		if len(pw) != 12 {
			t.Fatalf("RandomPassword2(12) = %q, want 12 characters", pw)
		}
		if strings.IndexByte(FIRSTCHARS, pw[0]) < 0 {
			t.Fatalf("RandomPassword2(12) = %q starts with %q, which is not in FIRSTCHARS", pw, pw[0])
		}
		if !MeetComplxity(pw) {
			t.Fatalf("RandomPassword2(12) = %q does not meet the complexity rules", pw)
		}
	}
}

// Values have to differ between calls.
func TestRandomPasswordIsNotRepeated(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		pw := RandomPassword2(12)
		if seen[pw] {
			t.Fatalf("RandomPassword2(12) repeated %q", pw)
		}
		seen[pw] = true
	}
}

// GeneratePassword derives a sha512-crypt hash with a fresh salt each time.
func TestGeneratePasswordUsesFreshSalt(t *testing.T) {
	first, err := GeneratePassword("hunter2")
	if err != nil {
		t.Fatalf("GeneratePassword: %v", err)
	}
	second, err := GeneratePassword("hunter2")
	if err != nil {
		t.Fatalf("GeneratePassword: %v", err)
	}
	if first == second {
		t.Errorf("GeneratePassword produced the same hash twice, so the salt did not change: %q", first)
	}
	if !strings.HasPrefix(first, "$6$") {
		t.Errorf("GeneratePassword = %q, want a $6$ hash", first)
	}
}

// The corrupted entry in the weak password list is split back into the two
// entries it merged.
func TestWeakPasswordsHasNoMergedEntry(t *testing.T) {
	for _, w := range WEAK_PASSWORDS {
		if strings.ContainsAny(w, "，、") {
			t.Errorf("weak password list contains a merged entry %q", w)
		}
	}
	if !containsString(WEAK_PASSWORDS, "root123#") {
		t.Error("weak password list does not contain \"root123#\"")
	}
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func TestMeetComplxity(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123456", false},
		{"123abcABC!@#", true},
		{"123abcABC-@=", true},
	}
	for _, c := range cases {
		if c.want != MeetComplxity(c.in) {
			t.Errorf("%s != %v", c.in, c.want)
		}
	}
}
