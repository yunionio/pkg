package filterclause

import "testing"

func TestParseFilterClause(t *testing.T) {
	for _, c := range []string{
		"abc.in(1,2,3)",
		"test.equals(1)",
	} {
		fc := ParseFilterClause(c)
		t.Logf("%s => %s", c, fc.String())
	}
}
