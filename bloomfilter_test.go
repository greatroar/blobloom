// Copyright 2020 the Blobloom authors
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
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	t.Parallel()

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
		assert.GreaterOrEqual(t, f.NumBits(), config.nbits)
		assert.LessOrEqual(t, f.NumBits(), config.nbits+BlockBits)
		assert.True(t, f.Empty())

		for _, k := range keys {
			assert.False(t, f.Has(k))
		}
		for _, k := range keys {
			f.Add(k)
		}
		assert.False(t, f.Empty())
		for _, k := range keys {
			assert.True(t, f.Has(k))
		}

		f.Clear()
		assert.True(t, f.Empty())
		for _, k := range keys {
			assert.False(t, f.Has(k))
		}

		f.Fill()
		assert.False(t, f.Empty())
		for _, k := range keys {
			assert.True(t, f.Has(k))
		}
	}
}

func TestUse(t *testing.T) {
	t.Parallel()

	const n = 100000

	// For FPR = .01, n = 100000, the optimal number of bits is 958505.84
	// for a standard Bloom filter.
	f := NewOptimized(Config{
		Capacity: n,
		FPRate:   .01,
	})
	if f.NumBits() < 958506 {
		t.Fatalf("bloom filter with %d bits too small", f.NumBits())
	}

	t.Logf("k = %d; m/n = %d/%d = %.3f",
		f.k, f.NumBits(), n, float64(f.NumBits())/n)

	// Generate random hash values for n keys. Pretend the keys are all distinct,
	// even if the hashes are not.
	r := rand.New(rand.NewSource(0xb1007))
	hashes := make([]uint64, n)
	for i := range hashes {
		hashes[i] = r.Uint64()
	}

	for _, h := range hashes {
		f.Add(h)
	}

	for _, h := range hashes {
		if !f.Has(h) {
			t.Errorf("%032x added to Bloom filter but not found", h)
		}
	}

	// Generate some more random hashes to get a sense of the FPR.
	// Pretend these represent unique keys, distinct from the ones we added.
	const nTest = 10000
	fp := 0
	for i := 0; i < nTest; i++ {
		if f.Has(r.Uint64()) {
			fp++
		}
	}

	fpr := float64(fp) / nTest
	assert.Less(t, fpr, .02)
	t.Logf("FPR = %.5f\n", fpr)
}

// Test robustness against 32-bit hash functions.
func TestHash32(t *testing.T) {
	t.Parallel()

	const n = 400

	f := NewOptimized(Config{
		Capacity: n,
		FPRate:   .01,
	})

	r := rand.New(rand.NewSource(32))

	for i := 0; i < n; i++ {
		f.Add(uint64(r.Uint32()))
	}

	const nrounds = 8
	fp := 0
	for i := n; i < nrounds*n; i++ {
		if f.Has(uint64(r.Uint32())) {
			fp++
		}
	}

	fprate := float64(fp) / (nrounds * n)
	t.Logf("FP rate = %.2f%%", 100*fprate)
	assert.LessOrEqual(t, fprate, .1)
}

func TestDoubleHashing(t *testing.T) {
	t.Parallel()

	var h1, h2 uint32 = 0, 0

	for i := 0; i < 20; i++ {
		h1, h2 = doublehash(h1, h2, i)
		assert.NotEqual(t, h2, 0)
	}
}

func TestReducerange(t *testing.T) {
	t.Parallel()

	for i := 0; i < 40000; i++ {
		m := rand.Uint32()
		j := reducerange(rand.Uint32(), m)
		if m == 0 {
			assert.Equal(t, j, 0)
		}
		assert.Less(t, j, m)
	}
}

func TestCardinality(t *testing.T) {
	t.Parallel()

	const cap = 1e4
	f := NewOptimized(Config{
		Capacity: cap,
		FPRate:   .0015,
	})

	assert.EqualValues(t, 0, f.Cardinality())

	r := rand.New(rand.NewSource(0x81feae2b))

	var sumN, sumNhat float64
	for n := 1.0; n <= 5*cap; n++ {
		f.Add(r.Uint64())

		nhat := f.Cardinality()
		assert.InDelta(t, 1, nhat/float64(n), 0.09)

		sumN += n
		sumNhat += nhat
		if int(n)%cap == 0 {
			// On average, we want to be less than a percent off.
			assert.InDelta(t, 1, sumNhat/sumN, 0.008)
		}
	}
}

func TestCardinalityFull(t *testing.T) {
	t.Parallel()

	f := New(BlockBits, 2)
	for i := range f.b {
		for j := range f.b[i] {
			f.b[i][j] = ^uint32(0)
		}
	}

	assert.Equal(t, math.Inf(+1), f.Cardinality())
}

func TestIntersect(t *testing.T) {
	t.Parallel()

	const n uint64 = 1e4
	const seed = 0x5544332211
	hashes := randomU64(int(n), seed)

	f := NewOptimized(Config{Capacity: n, FPRate: 1e-3})
	g := NewOptimized(Config{Capacity: n, FPRate: 1e-3})
	i := NewOptimized(Config{Capacity: n, FPRate: 1e-3})

	for _, h := range hashes[:n/3] {
		f.Add(h)
	}
	for _, h := range hashes[n/3 : 2*n/3] {
		f.Add(h)
		g.Add(h)
		i.Add(h)
	}
	for _, h := range hashes[n/3:] {
		g.Add(h)
	}

	expectFPR := math.Min(f.FPRate(n), g.FPRate(n))

	f.Intersect(g)
	assert.NotEqual(t, i, g)

	for _, h := range hashes[n/3 : 2*n/3] {
		assert.True(t, f.Has(h))
	}

	var fp uint64
	for _, h := range hashes {
		if f.Has(h) && !i.Has(h) {
			fp++
		}
	}
	actualFPR := float64(fp) / float64(n)
	assert.Less(t, actualFPR, 2*expectFPR)
	t.Logf("FPR = %f", actualFPR)

	assert.Panics(t, func() { f.Intersect(New(f.NumBits(), 9)) })
	assert.Panics(t, func() { f.Union(New(n+BlockBits, f.k)) })
}

func TestUnion(t *testing.T) {
	t.Parallel()

	const n = 1e5
	hashes := randomU64(n, 0xa6e98fb)

	f := New(n, 5)
	g := New(n, 5)
	u := New(n, 5)

	for _, h := range hashes[:n/2] {
		f.Add(h)
		u.Add(h)
	}
	for _, h := range hashes[n/2:] {
		g.Add(h)
		u.Add(h)
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

func TestUnionSmall(t *testing.T) {
	t.Parallel()

	f := New(BlockBits, 2)
	g := New(BlockBits, 2)

	g.Add(42)

	f.Union(g)
	assert.True(t, f.Has(42))
}

// This test ensures that the switch from 64-bit to 32-bit words did not
// alter the little-endian serialization of blocks.
func TestBlockLayout(t *testing.T) {
	t.Parallel()

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

// This test ensures that TestLocations() has same behavior with Has()
func TestLocations(t *testing.T) {
	t.Parallel()

	const n = 100000

	// For FPR = .01, n = 100000, the optimal number of bits is 958505.84
	// for a standard Bloom filter.
	f := NewOptimized(Config{
		Capacity: n,
		FPRate:   .01,
	})
	if f.NumBits() < 958506 {
		t.Fatalf("bloom filter with %d bits too small", f.NumBits())
	}

	t.Logf("k = %d; m/n = %d/%d = %.3f",
		f.k, f.NumBits(), n, float64(f.NumBits())/n)

	// Generate random hash values for n keys. Pretend the keys are all distinct,
	// even if the hashes are not.
	r := rand.New(rand.NewSource(0xb1007))
	hashes := make([]uint64, n)
	for i := range hashes {
		hashes[i] = r.Uint64()
	}

	for _, h := range hashes {
		f.Add(h)
	}

	for _, h := range hashes {
		ret1 := f.Has(h)
		locs := Locations(h, f.K())
		ret2 := f.TestLocations(locs)
		assert.Equal(t, ret1, ret2)
	}

	// Generate some more random hashes to get a sense of the FPR.
	// Pretend these represent unique keys, distinct from the ones we added.
	const nTest = 10000
	fp1 := 0
	fp2 := 0
	for i := 0; i < nTest; i++ {
		h := r.Uint64()
		if f.Has(h) {
			fp1++
		}

		locs := Locations(h, f.K())
		if f.TestLocations(locs) {
			fp2++
		}
	}

	fpr1 := float64(fp1) / nTest
	fpr2 := float64(fp2) / nTest
	assert.Equal(t, fpr1, fpr2)
}

func TestMarshal(t *testing.T) {
	t.Parallel()

	const n = 100000

	// For FPR = .01, n = 100000, the optimal number of bits is 958505.84
	// for a standard Bloom filter.
	f := NewOptimized(Config{
		Capacity: n,
		FPRate:   .01,
	})
	if f.NumBits() < 958506 {
		t.Fatalf("bloom filter with %d bits too small", f.NumBits())
	}

	t.Logf("k = %d; m/n = %d/%d = %.3f",
		f.k, f.NumBits(), n, float64(f.NumBits())/n)

	// Generate random hash values for n keys. Pretend the keys are all distinct,
	// even if the hashes are not.
	r := rand.New(rand.NewSource(0xb1007))
	hashes := make([]uint64, n)
	for i := range hashes {
		hashes[i] = r.Uint64()
	}

	for _, h := range hashes {
		f.Add(h)
	}

	for _, h := range hashes {
		ret1 := f.Has(h)
		assert.True(t, ret1)
	}

	data, err := f.MarshalJSON()
	assert.NoError(t, err)

	f2 := &Filter{}
	f2.UnmarshalJSON(data)

	for _, h := range hashes {
		ret1 := f.Has(h)
		assert.True(t, ret1)
	}
}
