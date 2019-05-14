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
