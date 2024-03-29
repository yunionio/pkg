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
	"strings"
	"testing"
)

func TestParseStructFieldJsonInfo_Name(t *testing.T) {
	type T struct {
		FiNameTag       int `name:"heck" json:"other" validate:"name=heck,marshal_name=heck"`
		FiNameTagIgnore int `name:"heck" json:"-"     validate:"name=heck,marshal_name=heck"`

		FiCamel  int `validate:"name=fi_camel,marshal_name=fi_camel"`
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

func BenchmarkFetchStructFieldValueSet(b *testing.B) {
	type GuestIp struct {
		GuestIpStart string `width:"16" charset:"ascii" nullable:"false" list:"user" update:"user" create:"required"`
		GuestIpEnd   string `width:"16" charset:"ascii" nullable:"false" list:"user" update:"user" create:"required"`
		GuestIpMask  int8   `nullable:"false" list:"user" update:"user" create:"required"`
	}
	type Network struct {
		GuestIp
		VlanId int    `nullable:"false" default:"1" list:"user" update:"user" create:"optional"`
		WireId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required"`
	}
	j := Network{
		GuestIp: GuestIp{
			GuestIpStart: "10.168.10.1",
			GuestIpEnd:   "10.168.10.244",
			GuestIpMask:  24,
		},
		VlanId: 123,
		WireId: "8324234723a",
	}
	v := reflect.ValueOf(j)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FetchStructFieldValueSet(v)
	}
}

func TestGetStructFieldIndexes(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Struct1 struct {
		Embeded
		Prop1 string `json:"prop1"`
	}
	type Struct2 struct {
		Embeded
		Prop2 string `json:"prop2"`
	}
	type TopStruct struct {
		Struct1
		Struct2
	}
	s := TopStruct{}
	set := FetchStructFieldValueSet(reflect.ValueOf(s))
	indexes := set.GetStructFieldIndexes("name")
	t.Logf("%v", indexes)
}

func TestParseFieldJsonInfo(t *testing.T) {
	cases := []struct {
		Name string
		Tag  string
		Want string
	}{
		{
			Name: "DBInstanceId",
			Tag:  `name:"dbinstance_id"`,
			Want: "dbinstance_id",
		},
		{
			Name: "DBInstanceId",
			Tag:  `json:"dbinstance_id"`,
			Want: "dbinstance_id",
		},
		{
			Name: "DBInstanceId",
			Tag:  ``,
			Want: "db_instance_id",
		},
	}

	for _, c := range cases {
		info := ParseFieldJsonInfo(c.Name, reflect.StructTag(c.Tag))
		if info.Name != c.Want {
			t.Errorf("TestParseFieldJsonInfo want %s got %s", c.Want, info.Name)
		}
	}
}

func TestOverrideStructTagsCompond(t *testing.T) {
	type StatusBase struct {
		Status string `default:"init"`
	}
	type EnabledBase struct {
		Enabled *bool `default:"false"`
	}
	type Compond struct {
		StatusBase `default:"offline"`
		EnabledBase
	}
	type TopStruct struct {
		Compond `"status->default":"online" "enabled->default":"true"`
	}
	cases := []struct {
		Object interface{}
		Want   map[string]map[string]string
	}{
		{
			StatusBase{},
			map[string]map[string]string{
				"status": map[string]string{
					"default": "init",
				},
			},
		},
		{
			EnabledBase{},
			map[string]map[string]string{
				"enabled": map[string]string{
					"default": "false",
				},
			},
		},
		{
			Compond{},
			map[string]map[string]string{
				"status": map[string]string{
					"default": "offline",
				},
				"enabled": map[string]string{
					"default": "false",
				},
			},
		},
		{
			TopStruct{},
			map[string]map[string]string{
				"status": map[string]string{
					"default": "online",
				},
				"enabled": map[string]string{
					"default": "true",
				},
			},
		},
	}
	for _, c := range cases {
		set := FetchStructFieldValueSet(reflect.ValueOf(c.Object))
		got := make(map[string]map[string]string)
		for _, s := range set {
			got[s.Info.MarshalName()] = s.Info.Tags
		}
		if !reflect.DeepEqual(got, c.Want) {
			t.Errorf("Got: %s Want: %s", got, c.Want)
		}
	}
}

func TestOverrideStructTags(t *testing.T) {
	type Embeded struct {
		Name string `json:"name" update:"user"`
	}
	type Struct1 struct {
		Embeded `update:"admin" create:"required" default:"emily"`
	}
	type Struct2 struct {
		Embeded `update:"domain" create:"optional"`
	}
	type TopStruct struct {
		Struct1 `create:"optional" default:""`
	}

	cases := []struct {
		Object interface{}
		Want   map[string]string
	}{
		{
			Embeded{},
			map[string]string{
				"json":   "name",
				"update": "user",
			},
		},
		{
			Struct1{},
			map[string]string{
				"json":    "name",
				"update":  "admin",
				"create":  "required",
				"default": "emily",
			},
		},
		{
			Struct2{},
			map[string]string{
				"json":   "name",
				"update": "domain",
				"create": "optional",
			},
		},
		{
			TopStruct{},
			map[string]string{
				"json":    "name",
				"update":  "admin",
				"create":  "optional",
				"default": "",
			},
		},
	}
	for _, c := range cases {
		set := FetchStructFieldValueSet(reflect.ValueOf(c.Object))
		if !reflect.DeepEqual(set[0].Info.Tags, c.Want) {
			t.Errorf("Got: %s Want: %s", set[0].Info.Tags, c.Want)
		}
	}
}

func TestExpandAmbiguousPrefix(t *testing.T) {
	type Embeded struct {
		Provider      string `json:"provider"`
		Cloudregion   string `json:"cloudregion"`
		CloudregionId string `json:"cloudregion_id" yunion-deprecated-by:"cloudregion"`
	}
	type Struct1 struct {
		Embeded

		Name string `json:"name"`
	}
	type Struct2 struct {
		Embeded

		Value string `json:"value"`
	}
	type TopStruct struct {
		Struct1 `yunion-ambiguous-prefix:"vpc_"`
		Struct2
	}
	type TopStruct2 struct {
		Struct1
		Struct2
	}

	cases := []struct {
		Obj   interface{}
		Cases []struct {
			Key   string
			Count int
		}
	}{
		{
			Obj: TopStruct{},
			Cases: []struct {
				Key   string
				Count int
			}{
				{
					Key:   "provider",
					Count: 1,
				},
				{
					Key:   "cloudregion",
					Count: 1,
				},
				{
					Key:   "cloudregion_id",
					Count: 1,
				},
				{
					Key:   "vpc_provider",
					Count: 1,
				},
				{
					Key:   "vpc_cloudregion",
					Count: 1,
				},
				{
					Key:   "vpc_cloudregion_id",
					Count: 1,
				},
				{
					Key:   "name",
					Count: 1,
				},
				{
					Key:   "value",
					Count: 1,
				},
			},
		},
		{
			Obj: TopStruct2{},
			Cases: []struct {
				Key   string
				Count int
			}{
				{
					Key:   "provider",
					Count: 2,
				},
				{
					Key:   "cloudregion",
					Count: 2,
				},
				{
					Key:   "cloudregion_id",
					Count: 2,
				},
				{
					Key:   "vpc_provider",
					Count: 0,
				},
				{
					Key:   "vpc_cloudregion",
					Count: 0,
				},
				{
					Key:   "vpc_cloudregion_id",
					Count: 0,
				},
				{
					Key:   "name",
					Count: 1,
				},
				{
					Key:   "value",
					Count: 1,
				},
			},
		},
	}

	for _, c := range cases {
		fields := FetchStructFieldValueSet(reflect.ValueOf(c.Obj))
		for _, ci := range c.Cases {
			indexes := fields.GetStructFieldIndexes2(ci.Key, true)
			if len(indexes) != ci.Count {
				t.Errorf("key %s expect %d got %d", ci.Key, ci.Count, len(indexes))
			} else {
				t.Logf("key %s expect %d got %d", ci.Key, ci.Count, len(indexes))
			}
		}
	}
}

func TestEmbededStructPtr(t *testing.T) {
	type Embeded0 struct {
		Field0 string `json:"field0"`
	}

	type Embeded1 struct {
		Field1 string `json:"field1"`
	}

	type Embeded2 struct {
		Embeded0
		*Embeded1

		Field2 string `json:"field2"`
		Field3 string `json:"field3"`
	}

	type Struct1 struct {
		*Embeded2

		Name string `json:"name"`
	}

	type Struct2 struct {
		Embeded2

		Name string `json:"name"`
	}

	cases := []struct {
		Object interface{}
		Want   map[string][]int
	}{
		{
			Struct1{},
			map[string][]int{
				"field0": []int{0},
				"field1": []int{1},
				"field2": []int{2},
				"field3": []int{3},
				"name":   []int{4},
			},
		},
		{
			Struct2{},
			map[string][]int{
				"field0": []int{0},
				"field1": []int{1},
				"field2": []int{2},
				"field3": []int{3},
				"name":   []int{4},
			},
		},
	}
	for _, c := range cases {
		set := FetchStructFieldValueSet(reflect.ValueOf(c.Object))
		if !reflect.DeepEqual(set.GetStructFieldIndexesMap(), c.Want) {
			t.Errorf("Got: %v Want: %v", set.GetStructFieldIndexesMap(), c.Want)
		}
	}
}
