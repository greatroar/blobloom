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

package blobloom

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDumpLoad(t *testing.T) {
	f := New(12345, 6)
	r := rand.New(rand.NewSource(55))
	for i := 0; i < 100; i++ {
		f.Add(r.Uint64())
	}

	buf := new(bytes.Buffer)
	n, err := Dump(buf, f, "random bytes")
	require.NoError(t, err)
	assert.EqualValues(t, 26*64, n)

	l, err := NewLoader(buf)
	require.NoError(t, err)
	assert.Equal(t, "random bytes", l.Comment)

	g := New(12345, 6)
	g2, err := l.Load(g)
	require.NoError(t, err)
	assert.True(t, g == g2)
	assert.True(t, f.Equals(g))

	g2, err = l.Load(nil)
	assert.Nil(t, g2)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}
