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

package streamutils

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ulikunitz/xz"
)

func TestStreamPipe(t *testing.T) {
	for _, bufSize := range []int{
		324553,
		2312,
	} {
		seed := rand.New(rand.NewSource(int64(bufSize)))
		buf := make([]byte, bufSize)
		n, err := seed.Read(buf)
		if err != nil {
			t.Errorf("rand seed read fail %s", err)
			continue
		}
		t.Logf("rand read %d", n)

		inBuf := bytes.NewReader(buf[:n])
		outBuf := &bytes.Buffer{}
		stat, err := StreamPipe(inBuf, outBuf, true, nil)
		if err != nil {
			t.Errorf("Steampipe fail %s", err)
		} else {
			t.Logf("stat %#v", stat)
			if !reflect.DeepEqual(buf, outBuf.Bytes()) {
				t.Errorf("input != output")
			}
		}
	}
}

func TestStreamPipeXZ(t *testing.T) {
	for _, bufSize := range []int{
		324553,
		2312,
	} {
		seed := rand.New(rand.NewSource(int64(bufSize)))
		buf := make([]byte, bufSize)
		n, err := seed.Read(buf)
		if err != nil {
			t.Errorf("rand seed read fail %s", err)
			continue
		}
		t.Logf("rand read %d", n)

		xzBuf := &bytes.Buffer{}

		w, err := xz.NewWriter(xzBuf)
		if err != nil {
			t.Errorf("xz NewWriter fail %s", err)
			continue
		}

		n, err = io.WriteString(w, string(buf))
		if err != nil {
			t.Errorf("xz write fail %s", err)
			continue
		}
		t.Logf("xz to compress %d", n)

		err = w.Close()
		if err != nil {
			t.Errorf("xz write close fail %s", err)
			continue
		}

		xzBytes := xzBuf.Bytes()
		t.Logf("compressed %d", len(xzBytes))

		inBuf := bytes.NewReader(xzBytes)
		outBuf := &bytes.Buffer{}
		stat, err := StreamPipe(inBuf, outBuf, true, nil)
		if err != nil {
			t.Errorf("Steampipe fail %s", err)
		} else {
			t.Logf("stat %#v", stat)
			if !reflect.DeepEqual(buf, outBuf.Bytes()) {
				t.Errorf("input != output")
			}
		}
	}
}
