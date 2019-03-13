package reflectutils

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseStructFieldJsonInfo_Name(t *testing.T) {
	type T struct {
		FiCamel  int `validate:"name=,marshal_name=fi_camel"`
		FiIgnore int `json:"-" validate:"name=,marshal_name=fi_ignore"`
		FiDash   int `json:"-," validate:"name=-,marshal_name=-"`
		FiJson   int `json:"json" validate:"name=json,marshal_name=json"`
		FiName   int `json:"json" name:"name" validate:"name=name,marshal_name=name"`
	}

	v := T{}
	rt := reflect.TypeOf(v)
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		sfi := ParseStructFieldJsonInfo(sf)

		name := ""
		marshalName := ""
		for _, kv := range strings.Split(sfi.Tags["validate"], ",") {
			switch {
			case strings.HasPrefix(kv, "name="):
				name = kv[5:]
			case strings.HasPrefix(kv, "marshal_name="):
				marshalName = kv[13:]
			}
		}
		if sfi.Name != name {
			t.Errorf("field %s has Name %q, expecting %q", sf.Name, sfi.Name, name)
		}
		if sfi.MarshalName() != marshalName {
			t.Errorf("field %s has MarshalName %q, expecting %q",
				sf.Name, sfi.MarshalName(), marshalName)
		}
	}
}
