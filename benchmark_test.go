// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom

import (
	"encoding/binary"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests simulate a situation where SHA-256 hashes are stored in a
// Bloom filter, using the first eight bytes as the Bloom filter hashes.
// Reported speeds are sha256Size * number of SHA-256 hashes per second.

const sha256Size = 32

func benchmarkAdd(b *testing.B, nkeys int) {
	hash := make([]byte, nkeys*sha256Size)

	r := rand.New(rand.NewSource(98621))
	r.Read(hash)

	// We want to benchmark Add, not Optimize.
	bf := NewOptimized(Config{
		FPRate: .01,
		NKeys:  nkeys,
	})
	nhashes := bf.k
	nbits := bf.NBits()
	b.Logf("nhashes = %d, nbits = %d", nhashes, nbits)

	b.SetBytes(int64(len(hash)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bf = New(nbits, nhashes)
		addAllSha256(bf, hash)
	}
}

func BenchmarkAdd1e5(b *testing.B) { benchmarkAdd(b, 1e5) }
func BenchmarkAdd1e6(b *testing.B) { benchmarkAdd(b, 1e6) }
func BenchmarkAdd1e7(b *testing.B) { benchmarkAdd(b, 1e7) }

func benchmarkHasNegative(b *testing.B, nkeys int) {
	hash := make([]byte, nkeys*sha256Size)
	r := rand.New(rand.NewSource(0xa58a7))
	r.Read(hash)

	bf := NewOptimized(Config{
		FPRate: .01,
		NKeys:  nkeys,
	})
	addAllSha256(bf, hash)

	b.SetBytes(sha256Size)
	b.ResetTimer()

	h := make([]byte, sha256Size)
	fp := 0
	for i := 0; i < b.N; i++ {
		r.Read(h)
		if bf.Has64(binary.BigEndian.Uint64(h[:8])) {
			fp++
		}
	}

	b.StopTimer()

	if b.N < 10000 {
		return // Don't test the FPR in the trial runs.
	}
	fpr := float64(fp) / float64(b.N)
	assert.Less(b, fpr, .013)
}

func BenchmarkHasNegative1e5(b *testing.B) { benchmarkHasNegative(b, 1e5) }
func BenchmarkHasNegative1e6(b *testing.B) { benchmarkHasNegative(b, 1e6) }
func BenchmarkHasNegative1e7(b *testing.B) { benchmarkHasNegative(b, 1e7) }

func benchmarkHasPositive(b *testing.B, nkeys int) {
	hash := make([]byte, nkeys*sha256Size)
	r := rand.New(rand.NewSource(0xe5871))
	r.Read(hash)

	bf := NewOptimized(Config{
		FPRate: .01,
		NKeys:  nkeys,
	})
	addAllSha256(bf, hash)

	b.SetBytes(int64(len(hash)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		found := 0
		h := hash
		for len(h) > 0 {
			if bf.Has64(binary.BigEndian.Uint64(h[:8])) {
				found++
			}
			h = h[sha256Size:]
		}
		if found != nkeys {
			b.Fatal("some keys not found (or not distinct?)")
		}
	}
}

func BenchmarkHasPositive1e5(b *testing.B) { benchmarkHasPositive(b, 1e5) }
func BenchmarkHasPositive1e6(b *testing.B) { benchmarkHasPositive(b, 1e6) }
func BenchmarkHasPositive1e7(b *testing.B) { benchmarkHasPositive(b, 1e7) }

func addAllSha256(bf *Filter, hash []byte) {
	for len(hash) > 0 {
		bf.Add64(binary.BigEndian.Uint64(hash[:8]))
		hash = hash[sha256Size:]
	}
}
