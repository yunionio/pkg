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

func BenchmarkFetchStructFieldValueSet(b *testing.B) {
	type GuestIp struct {
		GuestIpStart string `width:"16" charset:"ascii" nullable:"false" list:"user" update:"user" create:"required"`
		GuestIpEnd   string `width:"16" charset:"ascii" nullable:"false" list:"user" update:"user" create:"required"`
		GuestIpMask  int8   `nullable:"false" list:"user" update:"user" create:"required"`
	}
	type Network struct {
		GuestIp
		VlanId int `nullable:"false" default:"1" list:"user" update:"user" create:"optional"`
		WireId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required"`
	}
	j := Network{
		GuestIp: GuestIp{
			GuestIpStart: "10.168.10.1",
			GuestIpEnd: "10.168.10.244",
			GuestIpMask: 24,
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
