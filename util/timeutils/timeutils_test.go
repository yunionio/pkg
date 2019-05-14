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

package timeutils

import (
	"testing"
	"time"
)

func TestTimeUtils(t *testing.T) {
	tm := time.Now().UTC()
	tmLocal := time.Now().UTC()

	t.Logf("isoTime: %s", IsoTime(tm))
	t.Logf("isoNoSecondTime: %s", IsoNoSecondTime(time.Time{}))
	t.Logf("FullIsoTime: %s", FullIsoTime(tm))
	t.Logf("mysqlTime: %s", MysqlTime(tm))
	t.Logf("CompactTime: %s", CompactTime(tm))
	t.Logf("ShortDate: %s", ShortDate(tm))
	t.Logf("Date: %s", DateStr(tm))
	t.Logf("RFC2882: %s", RFC2882Time(tm))
	t.Logf("ZStack: %s", ZStackTime(tmLocal))

	tm2, err := ParseTimeStr(IsoTime(tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	tm3, err := ParseTimeStr(MysqlTime(tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	tm4, err := ParseTimeStr(CompactTime(tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	tm5, err := ParseTimeStr(ZStackTime(tmLocal))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	if tm2 != tm3 || tm2 != tm4 {
		t.Errorf("Parse Iso time error! %s %s", tm, tm2)
	}

	if tmLocal.Sub(tm5) > 1*time.Second {
		t.Errorf("Parse ZStack time error! %s %s %s", tmLocal, tm5, tmLocal.Sub(tm5))
	}
}
