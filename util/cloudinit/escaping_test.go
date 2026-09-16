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
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", "'abc'"},
		{"", "''"},
		{"a b", "'a b'"},
		{"a;b", "'a;b'"},
		{"a'b", `'a'\''b'`},
		{"$HOME", "'$HOME'"},
		{"`id`", "'`id`'"},
	}
	for _, c := range cases {
		if got := shellQuote(c.in); got != c.want {
			t.Errorf("shellQuote(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestValidUserName(t *testing.T) {
	for _, ok := range []string{"root", "yunion", "a-b_c.d", "user123"} {
		if !validUserName(ok) {
			t.Errorf("validUserName(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "a\nb", "a\rb", "a\x00b", "a\x7fb"} {
		if validUserName(bad) {
			t.Errorf("validUserName(%q) = true, want false", bad)
		}
	}
}

// A name carrying shell metacharacters must only ever appear quoted.
func TestShellScriptsQuoteUserName(t *testing.T) {
	u := NewUser("x; id; #")
	u.HashedPasswd = "$6$abcdef"

	script := strings.Join(u.ShellScripts(), "\n")
	if !strings.Contains(script, `useradd -m 'x; id; #' || true`) {
		t.Errorf("useradd is not quoted:\n%s", script)
	}
	if !strings.Contains(script, `usermod -p '$6$abcdef' 'x; id; #'`) {
		t.Errorf("usermod is not quoted:\n%s", script)
	}
	// The unquoted form would let the name run its own command.
	if strings.Contains(script, "useradd -m x;") {
		t.Errorf("useradd carries the name unquoted:\n%s", script)
	}
	if strings.Contains(script, "chown -R x;") {
		t.Errorf("chown carries the name unquoted:\n%s", script)
	}
}

// A package name is an argument, not part of the command line.
func TestUserDataScriptQuotesPackages(t *testing.T) {
	conf := SCloudConfig{Packages: []string{"nginx; id"}}
	script := conf.UserDataScript()
	if !strings.Contains(script, `install -y 'nginx; id'`) {
		t.Errorf("package is not quoted:\n%s", script)
	}
	if strings.Contains(script, "install -y nginx; id") {
		t.Errorf("package carries its command unquoted:\n%s", script)
	}
}

// A line of the content that happens to match a fixed terminator must not end
// the heredoc early.
func TestHeredocTerminatorIsNotInContent(t *testing.T) {
	content := "line1\n_END\nrm -rf / #\n_END\n"
	joined := strings.Join(mkPutFileCmd("/etc/x", content, "", ""), "\n")

	if strings.Contains(joined, "<<'_END'") {
		t.Errorf("the heredoc uses a terminator that appears in the content:\n%s", joined)
	}
	if !strings.Contains(joined, content) {
		t.Errorf("the content was altered:\n%s", joined)
	}
	if !strings.Contains(joined, "<<'"+heredocPrefix) {
		t.Errorf("the heredoc is not quoted:\n%s", joined)
	}
}

// Two calls use different terminators, so content cannot be crafted against a
// fixed one.
func TestHeredocTerminatorVaries(t *testing.T) {
	first := heredocTerminator()
	second := heredocTerminator()
	if first == second {
		t.Errorf("the heredoc terminator is the same on every call: %q", first)
	}
}

// A user name with a control character is left out rather than written.
func TestUnusableUserNameIsSkipped(t *testing.T) {
	u := NewUser("bad\nname")
	if got := u.ShellScripts(); got != nil {
		t.Errorf("ShellScripts = %#v, want nil", got)
	}
	if got := u.PowerShellScripts(); got != nil {
		t.Errorf("PowerShellScripts = %#v, want nil", got)
	}
}

// PowerShell needs its own escaping: a backtick introduces an escape and $
// introduces a variable, even inside double quotes.
func TestPowerShellEscaping(t *testing.T) {
	u := NewUser(`a"b$c` + "`" + `d`)
	script := strings.Join(u.PowerShellScripts(), "\n")

	if strings.Contains(script, `"a"b$c`) {
		t.Errorf("the name is not escaped:\n%s", script)
	}
	if !strings.Contains(script, "a`\"b`$c``d") {
		t.Errorf("the escaped name is not present:\n%s", script)
	}
}

// net has no way to escape a quote, so a password carrying one is left out
// rather than written in a form that breaks out of it.
func TestPowerShellPasswordWithQuoteIsSkipped(t *testing.T) {
	u := NewUser("yunion")
	u.PlainTextPasswd = `pa"ss`
	script := strings.Join(u.PowerShellScripts(), "\n")
	if strings.Contains(script, "net user") {
		t.Errorf("the password was written despite carrying a quote:\n%s", script)
	}
	// The account is still created.
	if !strings.Contains(script, "New-LocalUser") {
		t.Errorf("the account is no longer created:\n%s", script)
	}
}

// Ordinary values produce the same shape of command as before, with quoting
// added.
func TestShellScriptsNormalInput(t *testing.T) {
	u := NewUser("yunion")
	u.SudoPolicy(USER_SUDO_NOPASSWD)
	u.SshKey("ssh-rsa AAAA")

	script := strings.Join(u.ShellScripts(), "\n")
	for _, want := range []string{
		"useradd -m 'yunion' || true",
		"chown -R 'yunion':'yunion' '/home/yunion/.ssh'",
		"/etc/sudoers.d/yunion",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("%q missing from:\n%s", want, script)
		}
	}
}
