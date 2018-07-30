package tristate

import "testing"

func TestTriState(t *testing.T) {
	if True.Bool() {
		t.Logf("True == true")
	}
	if !False.Bool() {
		t.Logf("False == false")
	}
	if !None.Bool() {
		t.Logf("None == false")
	}

	var val TriState
	if val.IsNone() {
		t.Logf("val is None")
	}
}
