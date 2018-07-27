package seclib

import "testing"

func TestRandomPassword(t *testing.T) {
	t.Logf("%s", RandomPassword(12))
}
