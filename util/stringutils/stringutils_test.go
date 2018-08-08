package stringutils

import "testing"

func TestStringUtils(t *testing.T) {
	t.Logf("%s", UUID4())
	t.Logf("%s", Interface2String(nil))
	t.Logf("%s", Interface2String(2))
	t.Logf("%s", Interface2String("test string"))
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
