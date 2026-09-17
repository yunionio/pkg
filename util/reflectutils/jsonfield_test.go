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
		validate, _ := sfi.Tag("validate")
		for _, kv := range strings.Split(validate, ",") {
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
			got[s.Info.MarshalName()] = s.Info.TagMap()
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
		if !reflect.DeepEqual(set[0].Info.TagMap(), c.Want) {
			t.Errorf("Got: %s Want: %s", set[0].Info.TagMap(), c.Want)
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

func TestAliases(t *testing.T) {
	type Struct1 struct {
		Field1 string `json:"field1" alias:"field1_alias"`
	}
	type Struct2 struct {
		Field2 string `json:"field2" alias:"field2_alias"`
	}
	type TopStruct struct {
		Struct1
		Struct2
	}

	cases := []struct {
		val   interface{}
		name  string
		index int
	}{
		{
			val:   TopStruct{},
			name:  "field1",
			index: 0,
		},
		{
			val:   TopStruct{},
			name:  "field2",
			index: 1,
		},
		{
			val:   TopStruct{},
			name:  "field1_alias",
			index: 0,
		},
		{
			val:   TopStruct{},
			name:  "field2_alias",
			index: 1,
		},
	}

	for _, c := range cases {
		set := FetchStructFieldValueSet(reflect.ValueOf(c.val))
		got := set.GetStructFieldIndex(c.name)
		if got != c.index {
			t.Errorf("Got: %v Want: %v", got, c.index)
		}
	}
}

func TestNilEmbededStructPtrValue(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Outer struct {
		*Embeded
	}

	// the embedded pointer is nil, the field is enumerated out of a value
	// allocated on the side and reported through Parent
	o := &Outer{}
	set := FetchStructFieldValueSet(reflect.ValueOf(o).Elem())
	if len(set) != 1 {
		t.Fatalf("want 1 field, got %d", len(set))
	}
	if !set[0].Value.CanInterface() {
		t.Errorf("the field of a nil embedded struct should be readable")
	}
	if set[0].Parent == nil || !set[0].Parent.Field.IsNil() {
		t.Fatalf("want the nil embedded pointer to be reported as the parent")
	}

	// assigning Parent.Value to Parent.Field puts the enumerated fields back
	// on the struct, which is what a caller has to do before writing to them
	set[0].Parent.Field.Set(set[0].Parent.Value)
	if o.Embeded == nil {
		t.Fatalf("the embedded pointer should have been allocated")
	}
	set[0].Value.Set(reflect.ValueOf("x"))
	if o.Name != "x" {
		t.Errorf("want x got %q", o.Name)
	}

	// an embedded pointer that is already allocated needs no adoption
	o2 := &Outer{Embeded: &Embeded{}}
	val, ok := FindStructFieldValue(reflect.ValueOf(o2).Elem(), "name")
	if !ok {
		t.Fatalf("should find the field of an allocated embedded struct")
	}
	val.Set(reflect.ValueOf("y"))
	if o2.Name != "y" {
		t.Errorf("want y got %q", o2.Name)
	}
}

func TestFetchStructFieldValueSetForWriteUnaddressable(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Outer struct {
		*Embeded
		Other string `json:"other"`
	}

	// a non-addressable value can not be modified in place
	o := Outer{}
	set := FetchAllStructFieldValueSetForWrite(reflect.ValueOf(o))
	if len(set) != 2 {
		t.Fatalf("want 2 fields, got %d", len(set))
	}
	// the field of the nil embedded pointer is enumerated out of a value
	// allocated on the side, so only an ordinary field is checked here
	idxOther := set.GetStructFieldIndex("other")
	if idxOther < 0 {
		t.Fatalf("field other not found")
	}
	if set[idxOther].Value.CanSet() {
		t.Errorf("field other should not be settable")
	}
	if o.Embeded != nil {
		t.Errorf("the embedded pointer should not be allocated")
	}

	// an addressable value still gets its embedded pointer allocated
	o2 := Outer{}
	set2 := FetchStructFieldValueSetForWrite(reflect.ValueOf(&o2).Elem())
	if o2.Embeded == nil {
		t.Fatalf("the embedded pointer should have been allocated")
	}
	idx := set2.GetStructFieldIndex("name")
	if idx < 0 {
		t.Fatalf("field name not found")
	}
	if !set2[idx].Value.CanSet() {
		t.Fatalf("the field of an allocated embedded struct should be settable")
	}
	set2[idx].Value.Set(reflect.ValueOf("x"))
	if o2.Name != "x" {
		t.Errorf("want x got %q", o2.Name)
	}
}

func TestExpandAmbiguousPrefixCollision(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Struct1 struct {
		Embeded
	}
	type Struct2 struct {
		Embeded

		VpcName string `json:"vpc_name"`
	}
	type TopStruct struct {
		Struct1 `yunion-ambiguous-prefix:"vpc_"`
		Struct2
	}

	// the prefix would take over the name of Struct2.VpcName, so Struct1's
	// Name keeps its original name instead
	set := FetchStructFieldValueSet(reflect.ValueOf(TopStruct{}))
	cases := []struct {
		Key   string
		Count int
	}{
		{"vpc_name", 1},
		{"name", 2},
	}
	for _, c := range cases {
		indexes := set.GetStructFieldIndexes2(c.Key, true)
		if len(indexes) != c.Count {
			t.Errorf("key %s expect %d got %d", c.Key, c.Count, len(indexes))
		}
	}
}

func TestExpandAmbiguousPrefixFixedPoint(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Struct1 struct {
		Embeded
	}
	type Struct2 struct {
		Embeded
	}
	type Named struct {
		AName string `json:"a_name"`
	}
	type TopStruct struct {
		Struct1 `yunion-ambiguous-prefix:"a_"`
		Struct2 `yunion-ambiguous-prefix:"b_"`
		Named
	}

	// a_ is taken by Named.AName, so Struct1 keeps name and Struct2 is
	// expanded, leaving every name used by exactly one field
	set := FetchStructFieldValueSet(reflect.ValueOf(TopStruct{}))
	seen := make(map[string]int)
	for i := range set {
		seen[set[i].Info.MarshalName()]++
	}
	for k, c := range seen {
		if c != 1 {
			t.Errorf("key %s used by %d fields, want 1", k, c)
		}
	}
	if _, ok := seen["b_name"]; !ok {
		t.Errorf("want b_name to be expanded, got %v", seen)
	}
}

func TestExpandAmbiguousPrefixAliases(t *testing.T) {
	type Embeded struct {
		Name string `json:"name" alias:"the_name"`
	}
	type Struct1 struct {
		Embeded
	}
	type Struct2 struct {
		Embeded
	}
	type TopStruct struct {
		Struct1 `yunion-ambiguous-prefix:"a_"`
		Struct2 `yunion-ambiguous-prefix:"b_"`
	}

	set := FetchStructFieldValueSet(reflect.ValueOf(TopStruct{}))
	for _, name := range []string{"a_name", "b_name"} {
		if len(set.GetStructFieldIndexes2(name, true)) != 1 {
			t.Errorf("key %s not expanded", name)
		}
	}
	// aliases are expanded along with the name
	for _, name := range []string{"a_the_name", "b_the_name"} {
		if len(set.GetStructFieldIndexes2(name, false)) != 1 {
			t.Errorf("alias %s not expanded", name)
		}
	}
	if len(set.GetStructFieldIndexes2("the_name", false)) != 0 {
		t.Errorf("the plain alias should not match any field any more")
	}
}

func TestFetchStructFieldValueSetNonStruct(t *testing.T) {
	type T struct {
		Name string `json:"name"`
	}
	var nilPtr *T
	values := []reflect.Value{
		reflect.ValueOf(3),
		reflect.ValueOf("str"),
		reflect.ValueOf(nilPtr),
		reflect.ValueOf([]string{"a"}),
		reflect.Value{},
		reflect.ValueOf(nil),
	}
	fetchers := []struct {
		name string
		f    func(reflect.Value) SStructFieldValueSet
	}{
		{"FetchStructFieldValueSet", FetchStructFieldValueSet},
		{"FetchStructFieldValueSetForWrite", FetchStructFieldValueSetForWrite},
		{"FetchAllStructFieldValueSet", FetchAllStructFieldValueSet},
		{"FetchAllStructFieldValueSetForWrite", FetchAllStructFieldValueSetForWrite},
	}
	for _, fetcher := range fetchers {
		for _, v := range values {
			set := fetcher.f(v)
			if len(set) != 0 {
				t.Errorf("%s(%v) want no field, got %d", fetcher.name, v, len(set))
			}
		}
	}

	// a struct value is still enumerated
	if len(FetchStructFieldValueSet(reflect.ValueOf(T{}))) != 1 {
		t.Errorf("a struct value should still be enumerated")
	}
}

func TestTagAccessors(t *testing.T) {
	info := ParseFieldJsonInfo("Foo", reflect.StructTag(`json:"foo" name:"a-name" width:"36" nullable:"false"`))

	if v, ok := info.Tag("width"); !ok || v != "36" {
		t.Errorf("Tag(width) = %q, %v", v, ok)
	}
	if _, ok := info.Tag("nonexistent"); ok {
		t.Errorf("Tag should not report a tag the field does not have")
	}
	if v, ok := info.Tag("nullable"); !ok || v != "false" {
		t.Errorf("Tag(nullable) = %q, %v", v, ok)
	}

	// the map TagMap hands out belongs to the caller
	m := info.TagMap()
	if m["json"] != "foo" || m["name"] != "a-name" {
		t.Errorf("TagMap = %v", m)
	}
	delete(m, "width")
	m["injected"] = "x"
	if _, ok := info.Tag("width"); !ok {
		t.Errorf("TagMap should hand out a copy, modifying it changed the info")
	}
	if _, ok := info.Tag("injected"); ok {
		t.Errorf("TagMap should hand out a copy, modifying it changed the info")
	}
}

func TestTagMapDoesNotTouchTheCache(t *testing.T) {
	type T struct {
		Width string `width:"36" charset:"ascii"`
	}
	v := reflect.ValueOf(T{})

	set := FetchStructFieldValueSet(v)
	m := set[0].Info.TagMap()
	delete(m, "width")
	delete(m, "charset")
	m["injected"] = "x"

	set2 := FetchStructFieldValueSet(v)
	if w, ok := set2[0].Info.Tag("width"); !ok || w != "36" {
		t.Errorf("the tags of a later fetch were affected: width = %q, %v", w, ok)
	}
	if c, ok := set2[0].Info.Tag("charset"); !ok || c != "ascii" {
		t.Errorf("the tags of a later fetch were affected: charset = %q, %v", c, ok)
	}
	if _, ok := set2[0].Info.Tag("injected"); ok {
		t.Errorf("the tags of a later fetch were affected")
	}
}

// 经带 tag 的内嵌取到的字段会覆盖 tag，不能污染缓存里那份
func TestTagOverrideNotLeaking(t *testing.T) {
	type Embeded struct {
		Name  string `json:"name" update:"user"`
		Other string `json:"other" alias:"the_alias"`
	}
	type Tagged struct {
		Embeded `update:"admin" create:"required"`
	}
	type Plain struct {
		Embeded
	}

	for _, f := range FetchStructFieldValueSet(reflect.ValueOf(Tagged{})) {
		tags := f.Info.TagMap()
		t.Logf("tagged: %-6s tags=%v", f.Info.MarshalName(), tags)
	}
	byName := map[string]SStructFieldValue{}
	for _, f := range FetchStructFieldValueSet(reflect.ValueOf(Plain{})) {
		if _, ok := f.Info.Tag("create"); ok {
			t.Errorf("%s: create 泄漏进缓存了: %v", f.Info.MarshalName(), f.Info.TagMap())
		}
		byName[f.Info.FieldName] = f
	}
	if got, _ := byName["Name"].Info.Tag("update"); got != "user" {
		t.Errorf("Name: want update=user got %q", got)
	}
	// 读取 unexported 字段是包内测试的特权，外部只能读到 TagMap
	if got := byName["Other"].Info.aliases; len(got) != 1 || got[0] != "the_alias" {
		t.Errorf("Other: aliases 被改动了: %v", got)
	}
}

// 歧义前缀会改 Name/tags/aliases，同样不能污染缓存
func TestAmbiguousPrefixNotLeaking(t *testing.T) {
	type Embeded struct {
		Name string `json:"name" yunion-deprecated-by:"name2" alias:"the_name"`
	}
	type S1 struct{ Embeded }
	type S2 struct{ Embeded }
	type Prefixed struct {
		S1 `yunion-ambiguous-prefix:"a_"`
		S2 `yunion-ambiguous-prefix:"b_"`
	}
	type Plain struct{ Embeded }

	for _, f := range FetchStructFieldValueSet(reflect.ValueOf(Prefixed{})) {
		dep, _ := f.Info.Tag(TAG_DEPRECATED_BY)
		t.Logf("prefixed: %-8s dep=%-10q aliases=%v", f.Info.MarshalName(), dep, f.Info.aliases)
	}
	for _, f := range FetchStructFieldValueSet(reflect.ValueOf(Plain{})) {
		if f.Info.Name != "name" {
			t.Errorf("want the cached name, got %q", f.Info.Name)
		}
		if dep, _ := f.Info.Tag(TAG_DEPRECATED_BY); dep != "name2" {
			t.Errorf("deprecated-by 泄漏: %q", dep)
		}
		if got := f.Info.aliases; len(got) != 1 || got[0] != "the_name" {
			t.Errorf("aliases 泄漏: %v", got)
		}
	}
}

// 并发：共享缓存下必须无竞态、无串扰
func TestFieldInfoConcurrent(t *testing.T) {
	type Embeded struct {
		Name string `json:"name"`
	}
	type Tagged struct {
		Embeded `update:"admin" create:"required"`
	}
	type Plain struct{ Embeded }

	done := make(chan struct{})
	for i := 0; i < 8; i++ {
		go func(i int) {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 400; j++ {
				if i%2 == 0 {
					_ = FetchStructFieldValueSet(reflect.ValueOf(Tagged{}))
				} else {
					for _, f := range FetchStructFieldValueSet(reflect.ValueOf(Plain{})) {
						if _, ok := f.Info.Tag("create"); ok {
							t.Errorf("并发下 tag 覆盖泄漏进缓存")
							return
						}
					}
				}
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		<-done
	}
}

// 带 tag 覆盖的内嵌结构体：字段的 tag 被覆写，info 需要私有化
type BTInner struct {
	A string
	B string
}
type BTOuter struct {
	BTInner `update:"admin" create:"required" default:"emily"`
	C       string
}

func BenchmarkFetchStructFieldValueSetTagged(b *testing.B) {
	v := reflect.ValueOf(BTOuter{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FetchStructFieldValueSet(v)
	}
}

// 歧义前缀路径：字段被改名
type BAInner struct {
	Name string `json:"name"`
}
type BAS1 struct{ BAInner }
type BAS2 struct{ BAInner }
type BATop struct {
	BAS1 `yunion-ambiguous-prefix:"a_"`
	BAS2 `yunion-ambiguous-prefix:"b_"`
}

func BenchmarkFetchStructFieldValueSetAmbiguous(b *testing.B) {
	v := reflect.ValueOf(BATop{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FetchStructFieldValueSet(v)
	}
}
