package timeutils

import (
	"testing"
	"time"
)

func TestTimeUtils(t *testing.T) {
	t.Logf("isoTime: %s", IsoTime(nil))
	t.Logf("isoNoSecondTime: %s", IsoNoSecondTime(time.Time{}))
	t.Logf("FullIsoTime: %s", FullIsoTime(nil))
	t.Logf("mysqlTime: %s", MysqlTime(nil))
	t.Logf("CompactTime: %s", CompactTime(nil))
	t.Logf("ShortDate: %s", ShortDate(nil))
	t.Logf("Date: %s", DateStr(nil))
	t.Logf("RFC2882: %s", RFC2882Time(nil))

	tm := time.Now().UTC()
	tm2, err := ParseTimeStr(IsoTime(&tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	tm3, err := ParseTimeStr(MysqlTime(&tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	tm4, err := ParseTimeStr(CompactTime(&tm))
	if err != nil {
		t.Errorf("Parse time str error: %s", err)
	}
	if tm2 != tm3 || tm2 != tm4 {
		t.Errorf("Parse Iso time error! %s %s", tm, tm2)
	}
}
