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

package cloudinit

import (
	"encoding/base64"
	"testing"
)

// A "#cloud-config" document whose body is valid YAML but not a mapping
// cannot be turned into a config and must be reported as an error.
func TestParseUserDataRejectsNonMapping(t *testing.T) {
	inputs := []string{
		"#cloud-config\n",
		"#cloud-config\n\n",
		"#cloud-config\n- a\n- b\n",
		"#cloud-config\n42\n",
		"#cloud-config\n\"hello\"\n",
		"#cloud-config\ntrue\n",
		"#cloud-config\nnull\n",
	}
	for _, in := range inputs {
		conf, err := ParseUserData(in)
		if err == nil {
			t.Errorf("ParseUserData(%q) returned no error, got %v", in, conf)
		}
		if conf != nil {
			t.Errorf("ParseUserData(%q) returned a non-nil config together with an error", in)
		}
	}
}

// The same inputs must be handled through the base64 entry point.
func TestParseUserDataBase64RejectsNonMapping(t *testing.T) {
	for _, in := range []string{"#cloud-config\n", "#cloud-config\n- a\n- b\n"} {
		conf, err := ParseUserDataBase64(base64.StdEncoding.EncodeToString([]byte(in)))
		if err == nil {
			t.Errorf("ParseUserDataBase64(%q) returned no error, got %v", in, conf)
		}
		if conf != nil {
			t.Errorf("ParseUserDataBase64(%q) returned a non-nil config together with an error", in)
		}
	}
}

// A mapping document still parses, and the shell form is still produced.
func TestParseUserDataAcceptsMapping(t *testing.T) {
	conf, err := ParseUserData("#cloud-config\nusers:\n  - name: root\n")
	if err != nil {
		t.Fatalf("ParseUserData: %v", err)
	}
	if conf == nil {
		t.Fatal("ParseUserData returned a nil config")
	}
	if conf.UserDataScript() == "" {
		t.Error("UserDataScript returned an empty script for a valid config")
	}
}

// A nil config must render as empty rather than dereferencing the receiver.
func TestNilConfigRendersEmpty(t *testing.T) {
	var conf *SCloudConfig
	if got := conf.UserData(); got != "" {
		t.Errorf("nil UserData() = %q, want empty", got)
	}
	if got := conf.UserDataScript(); got != "" {
		t.Errorf("nil UserDataScript() = %q, want empty", got)
	}
	if got := conf.UserDataScriptBase64(); got != "" {
		t.Errorf("nil UserDataScriptBase64() = %q, want empty", got)
	}
}

