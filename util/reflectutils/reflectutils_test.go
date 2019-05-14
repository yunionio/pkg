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

package reflectutils

import (
	"reflect"
	"testing"
)

type BaseTestStruct struct {
	Id   string
	Name string
}

type TestStruct struct {
	BaseTestStruct
	Age int
}

/*
func TestFindStructFieldValue(t *testing.T) {
	test := TestStruct{}
	testValue := reflect.Indirect(reflect.ValueOf(&test))
	val := FetchStructFieldNameValueInterfaces(testValue)
	t.Logf("%s", val)

	test.Name = "test"
	val1, find := FindStructFieldInterface(testValue, "name")
	if find {
		t.Logf("name: %s", val1)
	} else {
		t.Errorf("name not found!")
	}
}
*/

func TestFillEmbededStructValue(t *testing.T) {
	type BaseStruct struct {
		Id   string
		Name string
		Age  int
	}

	type Layer1Struct struct {
		BaseStruct
	}

	type Layer2Struct struct {
		Layer1Struct
	}

	base := &BaseStruct{Id: "1234567890", Name: "Test", Age: 24}
	layer2 := &Layer2Struct{}

	baseValue := reflect.Indirect(reflect.ValueOf(base))
	layer2Value := reflect.Indirect(reflect.ValueOf(layer2))
	filled := FillEmbededStructValue(layer2Value, baseValue)
	t.Logf("filed: %#v %#v", filled, layer2)
}

func TestExpandInterface(t *testing.T) {
	values := ExpandInterface([]string{"123", "456", "789"})
	t.Logf("count = %d", len(values))
	for _, val := range values {
		t.Logf("%s", val)
	}
}

func TestSetStructFieldValue(t *testing.T) {
	type TestStruct struct {
		Name string
	}
	val := TestStruct{}

	target := "Test Target"
	if SetStructFieldValue(reflect.Indirect(reflect.ValueOf(&val)), "name", reflect.ValueOf(target)) {
		if val.Name != target {
			t.Errorf("Fail to SetStructFieldValue")
		}
	} else {
		t.Errorf("Fail to find struct field")
	}
}
