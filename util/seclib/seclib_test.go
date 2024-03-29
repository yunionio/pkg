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

package seclib

import (
	"math/rand"
	"testing"
	"time"
)

func TestRandomPassword(t *testing.T) {
	t.Logf("%s", RandomPassword(12))
}

func TestRandomPassword2(t *testing.T) {
	rand.Seed(time.Now().Unix())
	t.Logf("%s", RandomPassword2(12))
}

func TestMeetComplxity(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123456", false},
		{"123abcABC!@#", true},
		{"123abcABC-@=", true},
	}
	for _, c := range cases {
		if c.want != MeetComplxity(c.in) {
			t.Errorf("%s != %v", c.in, c.want)
		}
	}
}
