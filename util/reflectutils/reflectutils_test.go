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
	t.Logf("filed: %v %s", filled, layer2)
}

func TestExpandInterface(t *testing.T) {
	values := ExpandInterface([]string{"123", "456", "789"})
	t.Logf("count = %d", len(values))
	for _, val := range values {
		t.Logf("%s", val)
	}
}
