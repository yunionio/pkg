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

package gotypes

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type sA int

func (a sA) String() string {
	return strconv.Itoa(int(a))
}

func (a sA) IsZero() bool {
	return a == 0
}

type sB struct {
	name string
}

func (b sB) String() string {
	return fmt.Sprintf("name: %s", b.name)
}

func (b sB) IsZero() bool {
	return len(b.name) == 0
}

type sC struct {
	name string
}

func (c *sC) String() string {
	return fmt.Sprintf("name: %s", c.name)
}

func (c *sC) IsZero() bool {
	return len(c.name) == 0
}

type sD string


func TestDefaultNewSerializable(t *testing.T) {
	var (
		pA = reflect.TypeOf((*sA)(nil))
		pB = reflect.TypeOf((*sB)(nil))
		pC = reflect.TypeOf((*sC)(nil))
		pd = reflect.TypeOf((*sD)(nil))
		i interface{} = ""
	)

	for _, ty := range []reflect.Type{pA, pA.Elem()} {
		ser, err := defaultNewSerializable(ty)
		if err != nil {
			t.Fatal("For sA: err should be nil")
		}
		_, ok := ser.(*sA)
		if !ok {
			t.Fatal("For sA: return value should be converted to '*sA' type")
		}
	}

	for _, ty := range []reflect.Type{pB, pB.Elem()} {
		ser, err := defaultNewSerializable(ty)
		if err != nil {
			t.Fatal("For sB: err should be nil")
		}
		_, ok := ser.(*sB)
		if !ok {
			t.Fatal("For sB: return value should be converted to '*sB' type")
		}
	}

	for _, ty := range []reflect.Type{pC, pC.Elem()} {
		ser, err := defaultNewSerializable(ty)
		if err != nil {
			t.Fatal("For sC: err should be nil")
		}
		_, ok := ser.(*sC)
		if !ok {
			t.Fatal("For sC: return value should be converted to '*sC' type")
		}
	}

	_, err := defaultNewSerializable(pd)
	if err == nil {
		t.Fatal("For sD: err should not be nil")
	}

	_, err = defaultNewSerializable(reflect.TypeOf(i))
	if err == nil {
		t.Fatal("For interface{}: err should not be nil")
	}
}


