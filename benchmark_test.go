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

// Benchmarks for the basic operations live in the benchmarks/ subpackage.

package blobloom

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

// Baseline for BenchmarkAddSync.
func benchmarkAddLocked(b *testing.B, nbits uint64) {
	const nhashes = 22 // Large number of hashes to create collisions.

	var (
		f    = New(nbits, nhashes)
		mu   sync.Mutex
		seed uint32
	)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(int64(atomic.AddUint32(&seed, 1))))
		for pb.Next() {
			mu.Lock()
			f.Add(r.Uint64())
			mu.Unlock()
		}
	})
}

func BenchmarkAddLocked128kB(b *testing.B) { benchmarkAddLocked(b, 1<<20) }
func BenchmarkAddLocked1MB(b *testing.B)   { benchmarkAddLocked(b, 1<<23) }
func BenchmarkAddLocked16MB(b *testing.B)  { benchmarkAddLocked(b, 1<<27) }

func benchmarkAddSync(b *testing.B, nbits uint64) {
	const nhashes = 22 // Large number of hashes to create collisions.

	f := NewSync(nbits, nhashes)
	var seed uint32

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(int64(atomic.AddUint32(&seed, 1))))
		for pb.Next() {
			f.Add(r.Uint64())
		}
	})
}

func BenchmarkAddSync128kB(b *testing.B) { benchmarkAddSync(b, 1<<20) }
func BenchmarkAddSync1MB(b *testing.B)   { benchmarkAddSync(b, 1<<23) }
func BenchmarkAddSync16MB(b *testing.B)  { benchmarkAddSync(b, 1<<27) }

func BenchmarkCardinalityDense(b *testing.B) {
	f := New(1<<20, 2)
	for i := range f.b {
		for j := range f.b[i] {
			f.b[i][j] = rand.Uint32()
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		f.Cardinality()
	}
}

func BenchmarkCardinalitySparse(b *testing.B) {
	f := New(1<<20, 2)
	for i := 0; i < len(f.b); i += 2 {
		for _, j := range []int{4, 8, 13} {
			f.b[i][j] = rand.Uint32()
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		f.Cardinality()
	}
}

func BenchmarkOnescount(b *testing.B) {
	var blk block
	for i := range blk {
		blk[i] = rand.Uint32()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		onescount(&blk)
	}
}

func BenchmarkUnion(b *testing.B) {
	const n = 1e6

	var (
		cfg    = Config{Capacity: n, FPRate: 1e-5}
		f      = NewOptimized(cfg)
		g      = NewOptimized(cfg)
		fRef   = NewOptimized(cfg)
		gRef   = NewOptimized(cfg)
		hashes = randomU64(n, 0xcb6231119)
	)

	b.Logf("NumBits = %d", f.NumBits())

	for _, h := range hashes[:n/2] {
		fRef.Add(h)
	}
	for _, h := range hashes[n/2:] {
		gRef.Add(h)
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
