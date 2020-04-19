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
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	keys := randomU64(10000, 0x758e326)

	for _, config := range []struct {
		nbits   uint64
		nhashes int
	}{
		{1, 2},
		{1024, 4},
		{100, 3},
		{10000, 7},
		{1000000, 14},
	} {
		f := New(config.nbits, config.nhashes)
		assert.GreaterOrEqual(t, f.NBits(), config.nbits)
		assert.LessOrEqual(t, f.NBits(), config.nbits+512)

		for _, k := range keys {
			assert.False(t, f.Has64(k))
		}
		for _, k := range keys {
			f.Add64(k)
		}
		for _, k := range keys {
			assert.True(t, f.Has64(k))
		}

		f.Clear()
		for _, k := range keys {
			assert.False(t, f.Has64(k))
		}
	}
}

func TestUse(t *testing.T) {
	const n = 100000

	// For FPR = .01, n = 100000, the optimal number of bits is 958505.84
	// for a standard Bloom filter.
	f := NewOptimized(Config{
		FPRate: .01,
		NKeys:  n,
	})
	if f.NBits() < 958506 {
		t.Fatalf("bloom filter with %d bits too small", f.NBits())
	}

	t.Logf("k = %d; m/n = %d/%d = %.3f",
		f.k, f.NBits(), n, float64(f.NBits())/n)

	// Generate random hash values for n keys. Pretend the keys are all distinct,
	// even if the hashes are not.
	// Assume that 100k random SHA-256 values are all distinct.
	r := rand.New(rand.NewSource(0xb1007))
	hashes := make([]uint64, n)
	for i := range hashes {
		hashes[i] = r.Uint64()
	}

	for _, h := range hashes {
		f.Add64(h)
	}

	for _, h := range hashes {
		if !f.Has64(h) {
			t.Errorf("%032x added to Bloom filter but not found", h)
		}
	}

	// Generate some more random hashes to get a sense of the FPR.
	// Pretend these represent unique keys, distinct from the ones we added.
	const nTest = 10000
	fp := 0
	for i := 0; i < nTest; i++ {
		if f.Has64(r.Uint64()) {
			fp++
		}
	}

	fpr := float64(fp) / nTest
	assert.Less(t, fpr, .02)
	t.Logf("FPR = %.5f\n", fpr)
}

func TestDoubleHashing(t *testing.T) {
	var h1, h2 uint32 = 0, 0

	for i := 0; i < 20; i++ {
		h1, h2 = doublehash(h1, h2, i)
		assert.NotEqual(t, h2, 0)
	}
}

func TestReducerange(t *testing.T) {
	for i := 0; i < 40000; i++ {
		m := rand.Uint32()
		j := reducerange(rand.Uint32(), m)
		if m == 0 {
			assert.Equal(t, j, 0)
		}
		assert.Less(t, j, m)
	}
}

func TestAtomic(t *testing.T) {
	var (
		ch  = make(chan uint64)
		f   = New(1<<13, 2)
		ref = New(1<<13, 2)
	)

	go func() {
		r := rand.New(rand.NewSource(0xaeb15))
		for i := 0; i < 1e4; i++ {
			h := r.Uint64()
			ref.Add64(h)
			ch <- h
		}
		close(ch)
	}()

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			for h := range ch {
				f.AddAtomic64(h)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, ref, f)
}

func TestUnion(t *testing.T) {
	const n = 1e5
	hashes := randomU64(n, 0xa6e98fb)

	f := New(n, 5)
	g := New(n, 5)
	u := New(n, 5)

	for _, h := range hashes[:n/2] {
		f.Add64(h)
		u.Add64(h)
	}
	for _, h := range hashes[n/2:] {
		g.Add64(h)
		u.Add64(h)
	}

	assert.NotEqual(t, f, g)

	f.Union(g)
	assert.Equal(t, u, f)
	assert.NotEqual(t, u, g)

	g.Union(f)
	assert.Equal(t, u, g)

	assert.Panics(t, func() { f.Union(New(n, 4)) })
	assert.Panics(t, func() { f.Union(New(n+BlockBits, 5)) })
}

func randomU64(n int, seed int64) []uint64 {
	r := rand.New(rand.NewSource(seed))
	p := make([]uint64, n)
	for i := range p {
		p[i] = r.Uint64()
	}
	return p
}

// This test ensures that the switch from 64-bit to 32-bit words did not
// alter the little-endian serialization of blocks.
func TestBlockLayout(t *testing.T) {
	var b block
	b.setbit(0)
	b.setbit(1)
	b.setbit(111)
	b.setbit(499)

	assert.Equal(t, BlockBits, 8*binary.Size(b))

	h := sha256.New()
	binary.Write(h, binary.LittleEndian, b)
	expect := "aa7f8c411600fa387f0c10641eab428a7ed2f27a86171ac69f0e2087b2aa9140"
	assert.Equal(t, expect, hex.EncodeToString(h.Sum(nil)))
}
