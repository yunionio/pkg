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

package appctx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// The language tag comes straight from the request, so parsing it has to cope
// with anything a client can send.
var malformedLangValues = []string{
	"",
	"x",
	"-",
	"---",
	"!!!!",
	"en-",
	"-en",
	"zh-Hans-",
	"und-",
	strings.Repeat("-", 1000),
	strings.Repeat("a", 4096),
	strings.Repeat("en-", 500),
	"\x00",
	"\xff\xfe",
	"🙂",
	"a" + string(rune(0x7f)) + "b",
}

func TestWithLangDoesNotPanic(t *testing.T) {
	for _, lang := range malformedLangValues {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("WithLang(%q) panicked: %v", lang, r)
				}
			}()
			Lang(WithLang(context.Background(), lang))
		}()
	}
}

// The same, through each of the three sources WithRequestLang reads.
func TestWithRequestLangDoesNotPanic(t *testing.T) {
	for _, lang := range malformedLangValues {
		for _, source := range []string{"query", "header", "cookie"} {
			req := httptest.NewRequest("GET", "/", nil)
			switch source {
			case "query":
				req.URL.RawQuery = "lang=" + url.QueryEscape(lang)
			case "header":
				req.Header.Set(LangHeader, lang)
			case "cookie":
				req.AddCookie(&http.Cookie{Name: "lang", Value: lang})
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s %q: WithRequestLang panicked: %v", source, lang, r)
					}
				}()
				Lang(WithRequestLang(context.Background(), req))
			}()
		}
	}
}

// A well formed tag is still honoured.
func TestWithLangAcceptsValidTag(t *testing.T) {
	for _, lang := range []string{"en", "zh-CN", "fr", "de-DE"} {
		tag := Lang(WithLang(context.Background(), lang))
		if tag.IsRoot() {
			t.Errorf("WithLang(%q) produced the root tag", lang)
		}
	}
}
