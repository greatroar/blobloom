// Copyright 2023 the Blobloom authors
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

//go:build go1.18
// +build go1.18

package blobloom

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func FuzzLoader(f *testing.F) {
	const validHeader = "blobloom\x00\x00\x00\x00" +
		"\x00\x00\x00\x00\x00\x00\x00\x02" + // one block, two hashes
		"this is a valid zero-padded UTF-8 comment\x00\x00\x00"
	var zeroblock [64]byte

	f.Add(zeroblock[:])
	f.Add([]byte(validHeader + string(zeroblock[:])))

	f.Fuzz(func(t *testing.T, p []byte) {
		r := bytes.NewReader(p)
		l, err := NewLoader(r)

		switch {
		case err != nil:
			if l != nil {
				t.Error("l should be nil when err != nil")
			}
			return
		case l.nblocks == 0:
			t.Fatal("l.nblocks == 0")
		case l.nhashes == 0:
			t.Fatal("l.nhashes == 0")
		case strings.IndexByte(l.Comment, 0) != -1:
			t.Fatal("zero in comment")
		}

		// Prevent large allocations.
		const maxMem = 1 << 20
		if l.nblocks > maxMem/(BlockBits/8) {
			t.Skip()
		}

		f, err := l.Load(nil)
		if err == nil {
			if f == nil {
				t.Error("err == nil and f == nil")
			}
		} else {
			if f != nil {
				t.Error("f != nil and err != nil")
			}
			if err != io.ErrUnexpectedEOF && !strings.HasPrefix(err.Error(), "blobloom: ") {
				t.Fatal("unexpected error", err)
			}
		}
	})
}
