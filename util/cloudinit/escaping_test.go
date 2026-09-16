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
	"regexp"
	"strings"
	"testing"
)

// heredocRe picks the delimiter out of a generated heredoc command.
var heredocRe = regexp.MustCompile(`<<'([^']*)'`)

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
	for _, ok := range []string{"root", "yunion", "a-b_c.d", "user123", "A1"} {
		if !validUserName(ok) {
			t.Errorf("validUserName(%q) = false, want true", ok)
		}
	}
	// A name becomes a directory under /home and a file under sudoers.d, so a
	// slash or a ".." would move what is written; a control character would
	// split a line; and a leading "-" could be read as an option.
	for _, bad := range []string{
		"", ".", "..", "-x",
		"a\nb", "a\rb", "a\x00b", "a\x7fb",
		"foo/../bar", "a/b", "../x", "a/",
		"a;b", "a$b", "a b", "a`b", `a"b`, "a'b", "a|b", "a&b", "café",
	} {
		if validUserName(bad) {
			t.Errorf("validUserName(%q) = true, want false", bad)
		}
	}
}

func TestValidWritePath(t *testing.T) {
	for _, ok := range []string{"/etc/hosts", "/etc/ansible/hosts", "relative/path.txt", "a..b"} {
		if !validWritePath(ok) {
			t.Errorf("validWritePath(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{
		"", "..", "../x", "../../etc/cron.d/evil",
		"/etc/../etc/shadow", "a/../../b", "/a/b/..",
	} {
		if validWritePath(bad) {
			t.Errorf("validWritePath(%q) = true, want false", bad)
		}
	}
}

// A name carrying shell metacharacters cannot be turned into a user at all.
func TestNameWithMetacharactersIsRejected(t *testing.T) {
	for _, name := range []string{"x; id; #", "foo/../bar", "$(id)", "a b", "-x", "`id`"} {
		u := NewUser(name)
		if got := u.ShellScripts(); got != nil {
			t.Errorf("ShellScripts for %q = %#v, want nil", name, got)
		}
		if got := u.PowerShellScripts(); got != nil {
			t.Errorf("PowerShellScripts for %q = %#v, want nil", name, got)
		}
	}
}

// A name that is allowed is quoted wherever it becomes an argument.
func TestShellScriptsQuoteUserName(t *testing.T) {
	u := NewUser("a.b-c_d")
	u.HashedPasswd = "$6$abcdef"

	script := strings.Join(u.ShellScripts(), "\n")
	for _, want := range []string{
		`useradd -m 'a.b-c_d' || true`,
		`usermod -p '$6$abcdef' 'a.b-c_d'`,
		`chown -R 'a.b-c_d':'a.b-c_d' '/home/a.b-c_d/.ssh'`,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("%q missing from:\n%s", want, script)
		}
	}
}

// A write_file whose path walks up out of its directory is refused.
func TestWriteFilePathWalkingUpIsRefused(t *testing.T) {
	for _, p := range []string{"../../etc/cron.d/evil", "/etc/../etc/shadow", "a/../../b"} {
		wf := NewWriteFile(p, "payload", "", "", false)
		if got := wf.ShellScripts(); got != nil {
			t.Errorf("ShellScripts for %q = %#v, want nil", p, got)
		}
	}
	for _, p := range []string{"/etc/hosts", "/etc/ansible/hosts", "relative/path.txt"} {
		wf := NewWriteFile(p, "payload", "", "", false)
		if len(wf.ShellScripts()) == 0 {
			t.Errorf("ShellScripts for %q is empty", p)
		}
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
	first := heredocTerminator("")
	second := heredocTerminator("")
	if first == second {
		t.Errorf("the heredoc terminator is the same on every call: %q", first)
	}
}

// The terminator is checked against the content rather than assumed unique.
// utils.GenRequestId returns the empty string when the random source is
// unavailable, which would otherwise leave the fixed prefix as the terminator.
func TestUniqueTerminator(t *testing.T) {
	cases := []struct {
		term    string
		content string
		want    string
	}{
		{heredocPrefix, "plain content", heredocPrefix},
		{heredocPrefix, "a line with " + heredocPrefix + " in it", heredocPrefix + "_"},
		{heredocPrefix, heredocPrefix + "_\n" + heredocPrefix + "__\n", heredocPrefix + "___"},
		{"abc", "nothing to match", "abc"},
	}
	for _, c := range cases {
		got := uniqueTerminator(c.term, c.content)
		if got != c.want {
			t.Errorf("uniqueTerminator(%q, %q) = %q, want %q", c.term, c.content, got, c.want)
		}
		if strings.Contains(c.content, got) {
			t.Errorf("uniqueTerminator returned %q, which occurs in the content", got)
		}
	}
}

// Whatever the terminator turns out to be, it must not occur in the content
// and must not collide with one the content embeds.
func TestHeredocTerminatorSafeForContent(t *testing.T) {
	for _, content := range []string{
		"line1\n_END\nrm -rf / #\n_END\n",
		heredocPrefix,
		heredocPrefix + "abc",
		strings.Repeat(heredocPrefix+"_", 4),
	} {
		term := heredocTerminator(content)
		if term == "" {
			t.Fatal("the terminator is empty")
		}
		if strings.Contains(content, term) {
			t.Errorf("the terminator %q occurs in the content %q", term, content)
		}
		joined := strings.Join(mkPutFileCmd("/etc/x", content, "", ""), "\n")
		m := heredocRe.FindStringSubmatch(joined)
		if m == nil {
			t.Fatalf("no heredoc in: %s", joined)
		}
		if strings.Contains(content, m[1]) {
			t.Errorf("the terminator %q used for the command occurs in the content", m[1])
		}
		if n := strings.Count(joined, m[1]); n != 2 {
			t.Errorf("the terminator %q appears %d times, want 2 (open and close)", m[1], n)
		}
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

// Every PowerShell line is parsed by PowerShell first, so $ introduces a
// subexpression and a backtick an escape, even inside double quotes.
func TestPowerShellEscapesSubexpression(t *testing.T) {
	u := NewUser("yunion")
	u.PlainTextPasswd = "x$(Get-Process)"

	script := strings.Join(u.PowerShellScripts(), "\n")
	// Escaping $ leaves the "$(...)" characters in place, so the escape is
	// what has to be checked for, not the absence of the substring.
	if strings.Contains(script, `"x$(Get-Process)"`) {
		t.Errorf("the password was written as a live subexpression:\n%s", script)
	}
	if !strings.Contains(script, "x`$(Get-Process)") {
		t.Errorf("the escaped password is not present:\n%s", script)
	}
	if !strings.Contains(script, `net user "yunion" "x`+"`"+`$(Get-Process)"`) {
		t.Errorf("the password line is not as expected:\n%s", script)
	}
}

// A control character written literally would break the line it sits on.
func TestPowerShellEscapesControlCharacters(t *testing.T) {
	cases := []struct {
		password string
		escaped  string
	}{
		{"foo\ncalc.exe\nbar", "foo`ncalc.exe`nbar"},
		{"foo\rbar", "foo`rbar"},
		{"foo\tbar", "foo`tbar"},
		{"`whoami`", "``whoami``"},
		{`pa"ss`, "pa`\"ss"},
	}
	for _, c := range cases {
		u := NewUser("yunion")
		u.PlainTextPasswd = c.password
		script := strings.Join(u.PowerShellScripts(), "\n")
		if !strings.Contains(script, c.escaped) {
			t.Errorf("password %q: %q missing from:\n%s", c.password, c.escaped, script)
		}
		// The raw value must not survive anywhere in the script.
		for _, line := range strings.Split(script, "\n") {
			if strings.Contains(line, c.password) && !strings.Contains(line, c.escaped) {
				t.Errorf("password %q appears unescaped on: %s", c.password, line)
			}
		}
	}
}

// All four lines name the same account: escaping the name for some of them and
// not others would point them at different accounts.
func TestPowerShellUsesTheSameNameInEveryLine(t *testing.T) {
	u := NewUser("yunion")
	u.PlainTextPasswd = "pw"
	script := strings.Join(u.PowerShellScripts(), "\n")
	if got := strings.Count(script, `"yunion"`); got != 4 {
		t.Errorf("the name appears %d times, want 4:\n%s", got, script)
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
