// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	r := rand.New(rand.NewSource(0x758e326))
	keys := make([]uint64, 10000)
	for i := range keys {
		keys[i] = r.Uint64()
	}

	for _, config := range []struct {
		nbits, nhashes int
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
