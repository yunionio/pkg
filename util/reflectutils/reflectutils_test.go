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

func TestFindAnonymouStructPointer(t *testing.T) {
	type Struct1 struct {
		Name1 string
	}
	type Struct2 struct {
		Name2 string
		Struct1
	}
	type Struct3 struct {
		Name3 string
		Struct2
	}

	s3 := Struct3{}
	var t3 *Struct3
	var t2 *Struct2
	var t1 *Struct1

	err := FindAnonymouStructPointer(&s3, t3)
	if err == nil {
		t.Errorf("t3 is not a pointer to pointer, shoud fail")
	}
	err = FindAnonymouStructPointer(s3, &t3)
	if err == nil {
		t.Errorf("s3 is not pointer, should fail")
	}

	err = FindAnonymouStructPointer(&s3, &t3)
	if err != nil {
		t.Errorf("Fail to find Struct3 in Struct3")
	}
	if t3 != &s3 {
		t.Errorf("t3 != &s3")
	}

	err = FindAnonymouStructPointer(&s3, &t2)
	if err != nil {
		t.Errorf("Fail to find Struct2 in Struct3")
	}
	if t2 != &s3.Struct2 {
		t.Errorf("t2 != &s3.Struct2")
	}

	err = FindAnonymouStructPointer(&s3, &t1)
	if err != nil {
		t.Errorf("Fail to find Struct1 in Struct3")
	}
	if t1 != &s3.Struct1 {
		t.Errorf("t1 != &s3.Struct1")
	}
}

func BenchmarkGetStructFieldIndexes2(b *testing.B) {
	type T struct {
		Foo     string
		FooBar  string
		FFooBar string
		FFFF    string
	}
	obj := T{}
	objV := reflect.ValueOf(obj)
	fvset := FetchAllStructFieldValueSet(objV)

	b.Run("nonexist,strict=false", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fvset.GetStructFieldIndexes2("ether_super_real", false)
		}
	})
	b.Run("exist,strict=false", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fvset.GetStructFieldIndexes2("foo", false)
		}
	})
}

func TestSetStructFieldValueTypeMismatch(t *testing.T) {
	type TestStruct struct {
		Count int
	}
	val := TestStruct{}
	dataValue := reflect.Indirect(reflect.ValueOf(&val))

	if SetStructFieldValue(dataValue, "count", reflect.ValueOf("not an int")) {
		t.Errorf("expected setting an incompatible value to fail")
	}
	if SetStructFieldValue(dataValue, "count", reflect.Value{}) {
		t.Errorf("expected setting an invalid value to fail")
	}
	if val.Count != 0 {
		t.Errorf("field should be untouched, got %d", val.Count)
	}
	if !SetStructFieldValue(dataValue, "count", reflect.ValueOf(3)) {
		t.Errorf("expected setting a compatible value to succeed")
	}
	if val.Count != 3 {
		t.Errorf("want 3 got %d", val.Count)
	}
}

func TestFillEmbededStructValueUnexported(t *testing.T) {
	type unexported struct {
		Id   string
		Name string
	}
	type Exported struct {
		Age int
	}
	type Outer struct {
		unexported
		Exported
	}

	o := Outer{}
	if FillEmbededStructValue(reflect.ValueOf(&o).Elem(), reflect.ValueOf(unexported{Id: "1"})) {
		t.Errorf("should not fill an unexported embedded struct")
	}
	if FillEmbededStructValue(reflect.ValueOf(o), reflect.ValueOf(Exported{Age: 3})) {
		t.Errorf("should not fill through a value that can not be modified")
	}
	if o.Age != 0 {
		t.Errorf("an unaddressable value should not be modified, got %d", o.Age)
	}
	if !FillEmbededStructValue(reflect.ValueOf(&o).Elem(), reflect.ValueOf(Exported{Age: 3})) {
		t.Errorf("fail to fill an exported embedded struct")
	}
	if o.Age != 3 {
		t.Errorf("want 3 got %d", o.Age)
	}
	if FillEmbededStructValue(reflect.Value{}, reflect.ValueOf(Exported{})) {
		t.Errorf("should not fill an invalid container")
	}
}

func TestFindAnonymouStructPointerUnexported(t *testing.T) {
	type hidden struct {
		A string
	}
	type Outer struct {
		hidden
	}

	var hp *hidden
	if err := FindAnonymouStructPointer(&Outer{}, &hp); err == nil {
		t.Errorf("should not find an unexported embedded struct")
	}
	if hp != nil {
		t.Errorf("want a nil pointer, got %v", hp)
	}

	// an exported embedded struct is still found
	type Exported struct {
		B string
	}
	type Outer2 struct {
		Exported
	}
	var ep *Exported
	if err := FindAnonymouStructPointer(&Outer2{}, &ep); err != nil {
		t.Errorf("fail to find an exported embedded struct: %v", err)
	}
	if ep == nil {
		t.Fatalf("want a pointer to the embedded struct")
	}
}
