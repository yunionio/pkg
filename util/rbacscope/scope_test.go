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

package rbacscope

import "testing"

func TestString2Scope(t *testing.T) {
	cases := []struct {
		in   string
		want TRbacScope
	}{
		{"system", ScopeSystem},
		{"SYSTEM", ScopeSystem},
		{"domain", ScopeDomain},
		{"project", ScopeProject},
		{"user", ScopeUser},
		{"none", ScopeNone},
		{"NONE", ScopeNone},
		// Compatibility spelling kept for boolean settings.
		{"true", ScopeSystem},
		// Anything else falls back to the default used by String2Scope.
		{"", ScopeProject},
		{"garbage", ScopeProject},
	}
	for _, c := range cases {
		if got := String2Scope(c.in); got != c.want {
			t.Errorf("String2Scope(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestString2ScopeDefault(t *testing.T) {
	cases := []struct {
		in   string
		def  TRbacScope
		want TRbacScope
	}{
		// An explicit "none" is honoured regardless of the default ...
		{"none", ScopeDomain, ScopeNone},
		{"none", ScopeSystem, ScopeNone},
		// ... while an empty string means "use the caller's default".
		{"", ScopeDomain, ScopeDomain},
		{"", ScopeSystem, ScopeSystem},
		{"unknown", ScopeUser, ScopeUser},
	}
	for _, c := range cases {
		if got := String2ScopeDefault(c.in, c.def); got != c.want {
			t.Errorf("String2ScopeDefault(%q, %q) = %q, want %q", c.in, c.def, got, c.want)
		}
	}
}

func TestScopeOrdering(t *testing.T) {
	ordered := []TRbacScope{ScopeNone, ScopeUser, ScopeProject, ScopeDomain, ScopeSystem}
	for i := 1; i < len(ordered); i++ {
		lower, higher := ordered[i-1], ordered[i]
		if !higher.HigherThan(lower) {
			t.Errorf("%q should rank above %q", higher, lower)
		}
		if !higher.HigherEqual(lower) {
			t.Errorf("%q should rank at or above %q", higher, lower)
		}
		if lower.HigherEqual(higher) {
			t.Errorf("%q should not rank at or above %q", lower, higher)
		}
	}
	for _, s := range ordered {
		if !s.HigherEqual(s) {
			t.Errorf("%q should rank at or above itself", s)
		}
	}
}
