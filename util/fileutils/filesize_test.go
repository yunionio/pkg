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
