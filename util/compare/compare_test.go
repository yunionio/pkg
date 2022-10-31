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

	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/utils"
)

type LocalRes struct {
	Name       string
	ExternalId string
}

func (r LocalRes) GetExternalId() string {
	return r.ExternalId
}

type RemoteRes struct {
	Name     string
	GlobalId string
}

func (r RemoteRes) GetGlobalId() string {
	return r.GlobalId
}

type TestData struct {
	db      []LocalRes
	remote  []RemoteRes
	common  []string
	removed []string
	added   []string

	duplicated []string
}

func TestCompareSets2(t *testing.T) {
	data := []TestData{
		{
			db: []LocalRes{
				{Name: "1", ExternalId: "1"},
				{Name: "2", ExternalId: "2"},
				{Name: "3", ExternalId: "3"},
				{Name: "4", ExternalId: "4"},
				{Name: "5", ExternalId: "5"},
			},
			remote: []RemoteRes{
				{Name: "2", GlobalId: "2"},
				{Name: "4", GlobalId: "4"},
				{Name: "6", GlobalId: "6"},
			},
			common:     []string{"2", "4"},
			removed:    []string{"1", "3", "5"},
			added:      []string{"6"},
			duplicated: []string{},
		},
		{
			db: []LocalRes{
				{Name: "1", ExternalId: ""},
				{Name: "2", ExternalId: "2"},
				{Name: "3", ExternalId: "3"},
				{Name: "4", ExternalId: "4"},
				{Name: "5", ExternalId: "5"},
			},
			remote: []RemoteRes{
				{Name: "2", GlobalId: "2"},
				{Name: "4", GlobalId: "4"},
				{Name: "6", GlobalId: "6"},
				{Name: "1-1", GlobalId: "1"},
				{Name: "1-2", GlobalId: "1"},
				{Name: "1-3", GlobalId: "1"},
			},
			common:     []string{"2", "4"},
			removed:    []string{"3", "5"},
			added:      []string{"6"},
			duplicated: []string{"1"},
		},
		{
			db: []LocalRes{
				{Name: "1", ExternalId: ""},
				{Name: "2", ExternalId: "2"},
				{Name: "3", ExternalId: "3"},
				{Name: "4", ExternalId: "4"},
				{Name: "5", ExternalId: "5"},
				{Name: "7", ExternalId: ""},
			},
			remote: []RemoteRes{
				{Name: "2", GlobalId: "2"},
				{Name: "4", GlobalId: "4"},
				{Name: "6", GlobalId: "6"},
				{Name: "7-1", GlobalId: "7"},
				{Name: "7-2", GlobalId: "7"},
			},
			common:     []string{"2", "4"},
			removed:    []string{"3", "5"},
			added:      []string{"6"},
			duplicated: []string{"7"},
		},
	}
	for _, d := range data {
		removed := []LocalRes{}
		commondb := []LocalRes{}
		commonext := []RemoteRes{}
		added := []RemoteRes{}
		duplicated := map[string][]RemoteRes{}
		err := CompareSets2(d.db, d.remote, &removed, &commondb, &commonext, &added, &duplicated)
		if err != nil {
			t.Fatalf("%v", err)
		}
		for i := range removed {
			if !utils.IsInStringArray(removed[i].Name, d.removed) {
				t.Errorf("%s should be remove", removed[i].Name)
			}
		}
		for i := range commondb {
			if !utils.IsInStringArray(commondb[i].Name, d.common) {
				t.Errorf("%s should be common", commondb[i].Name)
			}
		}
		for i := range added {
			if !utils.IsInStringArray(added[i].Name, d.added) {
				t.Errorf("%s should be added", added[i].Name)
			}
		}
		for k := range duplicated {
			t.Logf("duplicated key %s value %#v", k, duplicated[k])
			if !utils.IsInStringArray(k, d.duplicated) {
				t.Errorf("%s should be duplicated", k)
			}
		}
	}
}

func TestCompareSetsDuplicate(t *testing.T) {
	local := []LocalRes{
		{Name: "1", ExternalId: "1"},
	}
	remote := []RemoteRes{
		{Name: "test-dup1", GlobalId: "dup-id"},
		{Name: "test-dup2", GlobalId: "dup-id"},
	}

	removed := []LocalRes{}
	commondb := []LocalRes{}
	commonext := []RemoteRes{}
	added := []RemoteRes{}
	err := CompareSets(local, remote, &removed, &commondb, &commonext, &added)
	if err == nil || errors.Cause(err) != errors.ErrDuplicateId {
		t.Fatalf("should be %v error but is %v", errors.ErrDuplicateId, err)
	}
	t.Logf("test duplicate error: %v", err)
}
