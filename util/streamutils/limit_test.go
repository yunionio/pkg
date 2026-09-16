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
	"testing"

	"github.com/ulikunitz/xz"

	"yunion.io/x/pkg/errors"
)

func xzCompress(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w, err := xz.NewWriter(&buf)
	if err != nil {
		t.Fatalf("xz.NewWriter: %v", err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatalf("xz write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("xz close: %v", err)
	}
	return buf.Bytes()
}

// A stream that expands to much more than its own size must stop at the
// limit rather than filling the writer.
func TestStreamPipeSizeLimit(t *testing.T) {
	data := bytes.Repeat([]byte("a"), 1<<20)
	compressed := xzCompress(t, data)
	if len(compressed) >= len(data)/100 {
		t.Fatalf("input did not compress as expected: %d -> %d", len(data), len(compressed))
	}

	// Without a limit the whole input is produced.
	out := &bytes.Buffer{}
	stat, err := StreamPipe(bytes.NewReader(compressed), out, true, nil)
	if err != nil {
		t.Fatalf("without a limit: %v", err)
	}
	if int(stat.Size) != len(data) {
		t.Errorf("without a limit: size = %d, want %d", stat.Size, len(data))
	}
	if out.Len() != len(data) {
		t.Errorf("without a limit: wrote %d bytes, want %d", out.Len(), len(data))
	}

	// With a limit below the expanded size the transfer stops and says so.
	const limit = 4096
	out = &bytes.Buffer{}
	_, err = StreamPipe(bytes.NewReader(compressed), out, true, nil, limit)
	if err == nil {
		t.Fatal("expected the size limit to be reported")
	}
	if cause := errors.Cause(err); cause != ErrSizeLimitExceeded {
		t.Errorf("error = %v, want ErrSizeLimitExceeded", cause)
	}
	if out.Len() > limit {
		t.Errorf("wrote %d bytes, past the limit of %d", out.Len(), limit)
	}
}

// A limit above the produced size does not interfere.
func TestStreamPipeSizeLimitNotReached(t *testing.T) {
	data := bytes.Repeat([]byte("b"), 4096)
	compressed := xzCompress(t, data)

	out := &bytes.Buffer{}
	if _, err := StreamPipe(bytes.NewReader(compressed), out, false, nil, int64(10*len(data))); err != nil {
		t.Fatalf("StreamPipe: %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Error("output differs from the input")
	}
}

// A limit of zero, and no limit at all, both mean "no limit".
func TestStreamPipeZeroLimitMeansUnlimited(t *testing.T) {
	data := bytes.Repeat([]byte("c"), 4096)
	compressed := xzCompress(t, data)

	for _, limit := range []int64{0, -1} {
		out := &bytes.Buffer{}
		if _, err := StreamPipe2(bytes.NewReader(compressed), out, false, nil, limit); err != nil {
			t.Fatalf("limit %d: %v", limit, err)
		}
		if !bytes.Equal(out.Bytes(), data) {
			t.Errorf("limit %d: output differs from the input", limit)
		}
	}
}

// oneByteReader hands out a single byte per Read call, which is allowed by
// io.Reader and is what a socket or pipe often does.
type oneByteReader struct {
	data []byte
	pos  int
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	p[0] = r.data[r.pos]
	r.pos++
	return 1, nil
}

// A stream that is handed over in pieces must still be recognised as xz.
func TestStreamPipeShortReads(t *testing.T) {
	data := bytes.Repeat([]byte("hello world "), 1000)
	compressed := xzCompress(t, data)

	out := &bytes.Buffer{}
	stat, err := StreamPipe(&oneByteReader{data: compressed}, out, false, nil)
	if err != nil {
		t.Fatalf("StreamPipe: %v", err)
	}
	if !bytes.Equal(out.Bytes(), data) {
		t.Errorf("output is %d bytes, want the %d bytes of xz payload", out.Len(), len(data))
	}
	if stat.Size != int64(len(data)) {
		t.Errorf("size = %d, want %d", stat.Size, len(data))
	}
}
