package compare

import (
	"testing"
)

type StringA string

type StringB string

func (a StringA) GetExternalId() string {
	return string(a)
}

func (b StringB) GetGlobalId() string {
	return string(b)
}

func TestCompareSets(t *testing.T) {
	arr1 := []StringA{
		StringA("1"),
		StringA("2"),
		StringA("3"),
		StringA("4"),
		StringA("5"),
	}
	arr2 := []StringB{
		StringB("2"),
		StringB("4"),
		StringB("6"),
	}
	removed := make([]StringA, 0)
	commonA := make([]StringA, 0)
	commonB := make([]StringB, 0)
	added := make([]StringB, 0)
	err := CompareSets(arr1, arr2, &removed, &commonA, &commonB, &added)
	if err != nil {
		t.Errorf("%s", err)
	} else {
		t.Logf("%s %s %s %s", removed, commonA, commonB, added)
	}
}
