// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom

import (
	"math/rand"
	"sync"
	"testing"
)

// Baseline for BenchmarkAddAtomic.
func benchmarkAddLocked(b *testing.B, nbits uint64) {
	const nhashes = 22

	f := New(nbits, nhashes)
	var mu sync.Mutex

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			mu.Lock()
			f.Add64(r.Uint64())
			mu.Unlock()
		}
	})
}

func BenchmarkAddLocked128kB(b *testing.B) { benchmarkAddLocked(b, 1<<20) }
func BenchmarkAddLocked1MB(b *testing.B)   { benchmarkAddLocked(b, 1<<23) }
func BenchmarkAddLocked16MB(b *testing.B)  { benchmarkAddLocked(b, 1<<27) }

func benchmarkAddAtomic(b *testing.B, nbits uint64) {
	const nhashes = 22 // Large number of hashes to create collisions.

	f := New(nbits, nhashes)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(rand.Int63()))
		for pb.Next() {
			f.AddAtomic64(r.Uint64())
		}
	})
}

func BenchmarkAddAtomic128kB(b *testing.B) { benchmarkAddAtomic(b, 1<<20) }
func BenchmarkAddAtomic1MB(b *testing.B)   { benchmarkAddAtomic(b, 1<<23) }
func BenchmarkAddAtomic16MB(b *testing.B)  { benchmarkAddAtomic(b, 1<<27) }

func BenchmarkUnion(b *testing.B) {
	const n = 1e6

	var (
		cfg    = Config{FPRate: 1e-5, NKeys: n}
		f      = NewOptimized(cfg)
		g      = NewOptimized(cfg)
		fRef   = NewOptimized(cfg)
		gRef   = NewOptimized(cfg)
		hashes = randomU64(n, 0xcb6231119)
	)

	for _, h := range hashes[:n/2] {
		fRef.Add64(h)
	}
	for _, h := range hashes[n/2:] {
		gRef.Add64(h)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f.Clear()
		f.Union(fRef)
		g.Clear()
		g.Union(gRef)
		b.StartTimer()

		f.Union(g)
	}
}
