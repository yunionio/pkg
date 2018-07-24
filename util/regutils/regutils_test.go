package regutils

import (
	"testing"
)

func TestIsFunction(t *testing.T) {
	cases := []struct {
		in  string
		out bool
	}{
		{"NOW()", true},
		{"test", false},
		{"CONCAT('123','456')", true},
	}
	for _, c := range cases {
		if MatchFunction(c.in) != c.out {
			t.Errorf("%s is a function: %v", c.in, c.out)
		}
	}

}
