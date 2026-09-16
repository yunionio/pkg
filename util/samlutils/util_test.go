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

package samlutils

import (
	"strings"
	"testing"
)

func TestSAMLFormEscapesValues(t *testing.T) {
	cases := []struct {
		name   string
		action string
		attrs  map[string]string
		// substrings that must NOT appear unescaped in the output
		reject []string
	}{
		{
			name:   "quote in value",
			action: "https://idp.example.com/sso",
			attrs:  map[string]string{"RelayState": `"><script>alert(1)</script>`},
			reject: []string{`"><script>`, `alert(1)</script>`},
		},
		{
			name:   "quote in action",
			action: `https://idp.example.com/sso"><script>alert(1)</script>`,
			attrs:  map[string]string{"SAMLResponse": "AAAA"},
			reject: []string{`"><script>`, `alert(1)</script>`},
		},
		{
			name:   "quote in attribute name",
			action: "https://idp.example.com/sso",
			attrs:  map[string]string{`x" onload="evil`: "v"},
			reject: []string{`onload="evil`},
		},
		{
			name:   "angle brackets and ampersand",
			action: "https://idp.example.com/sso",
			attrs:  map[string]string{"RelayState": "<b>a&b</b>"},
			reject: []string{"<b>", "a&b"},
		},
	}
	for _, c := range cases {
		out := SAMLForm(c.action, c.attrs)
		for _, r := range c.reject {
			if strings.Contains(out, r) {
				t.Errorf("%s: output still contains %q\n%s", c.name, r, out)
			}
		}
	}
}

// Plain values must survive unchanged apart from escaping.
func TestSAMLFormKeepsBenignValues(t *testing.T) {
	action := "https://idp.example.com/sso"
	attrs := map[string]string{
		"SAMLResponse": "PHNhbWxwOlJlc3BvbnNlPjwvc2FtbHA6UmVzcG9uc2U+",
		"RelayState":   "a-b_c.d~e",
	}
	out := SAMLForm(action, attrs)
	for k, v := range attrs {
		if !strings.Contains(out, `name="`+k+`" value="`+v+`"`) {
			t.Errorf("output does not carry %s=%s\n%s", k, v, out)
		}
	}
	if !strings.Contains(out, `action="`+action+`"`) {
		t.Errorf("output does not carry the action\n%s", out)
	}
}

// Attribute order must not vary between calls.
func TestSAMLFormIsDeterministic(t *testing.T) {
	attrs := map[string]string{"b": "2", "a": "1", "c": "3", "d": "4", "e": "5"}
	first := SAMLForm("https://idp.example.com/sso", attrs)
	for i := 0; i < 20; i++ {
		if got := SAMLForm("https://idp.example.com/sso", attrs); got != first {
			t.Fatalf("output differs between calls:\n%s\n%s", first, got)
		}
	}
}
