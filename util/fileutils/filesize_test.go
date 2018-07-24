package fileutils

import "testing"

func TestFileSize(t *testing.T) {
	var size int

	size, _ = parseSizeStr("2048", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeGb("10g", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeGb("2048", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeMb("10g", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeMb("2048", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeKb("10g", 'M', 1024)
	t.Logf("%d", size)
	size, _ = GetSizeKb("2048", 'M', 1024)
	t.Logf("%d", size)
}
