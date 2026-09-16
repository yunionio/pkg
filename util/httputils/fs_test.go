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

package httputils

import (
	"os"
	"path/filepath"
	"testing"
)

func setupDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "withindex"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "noindex"), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "withindex", "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return dir
}

func TestFileSystemOpenRegularFile(t *testing.T) {
	dir := setupDir(t)
	f, err := Dir(dir).Open("/withindex/index.html")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	if _, err := f.Stat(); err != nil {
		t.Errorf("Stat: %v", err)
	}
}

func TestFileSystemOpenDirWithIndex(t *testing.T) {
	dir := setupDir(t)
	f, err := Dir(dir).Open("/withindex")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if !s.IsDir() {
		t.Error("expected the directory itself to be returned")
	}
}

// A directory without an index file must not be served.
func TestFileSystemOpenDirWithoutIndex(t *testing.T) {
	dir := setupDir(t)
	f, err := Dir(dir).Open("/noindex")
	if err == nil {
		f.Close()
		t.Fatal("expected an error for a directory without an index file")
	}
}

// Opening a path that does not exist is reported, not returned as a file.
func TestFileSystemOpenMissing(t *testing.T) {
	dir := setupDir(t)
	if f, err := Dir(dir).Open("/does-not-exist"); err == nil {
		f.Close()
		t.Fatal("expected an error for a missing path")
	}
}

// Repeatedly opening and closing must not accumulate open descriptors.
func TestFileSystemOpenDoesNotLeakDescriptors(t *testing.T) {
	if _, err := os.Stat("/proc/self/fd"); err != nil {
		t.Skip("no /proc/self/fd on this platform")
	}
	dir := setupDir(t)
	fs := Dir(dir)

	countFDs := func() int {
		entries, err := os.ReadDir("/proc/self/fd")
		if err != nil {
			t.Fatalf("ReadDir: %v", err)
		}
		return len(entries)
	}

	check := func(path string) {
		for i := 0; i < 200; i++ {
			f, err := fs.Open(path)
			if err != nil {
				t.Fatalf("Open(%s): %v", path, err)
			}
			f.Close()
		}
	}

	for _, path := range []string{"/withindex", "/withindex/index.html"} {
		check(path) // warm up
		before := countFDs()
		check(path)
		if after := countFDs(); after > before {
			t.Errorf("Open(%s) leaked %d descriptors (%d -> %d)", path, after-before, before, after)
		}
	}
}
