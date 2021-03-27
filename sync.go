// Copyright 2021 the Blobloom authors
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

import "sync/atomic"

// A SyncFilter is a Bloom filter that can be accessed and updated
// by multiple goroutines concurrently.
//
// A SyncFilter behaves as a regular filter protected by a lock,
//
//	type SyncFilter struct {
//		Filter
//		lock sync.Mutex
//	}
//
// with each operation taking and releasing the lock.
type SyncFilter struct {
	b []block // Shards.
	k int     // Number of hash functions required.
}

func NewSync(nbits uint64, nhashes int) *SyncFilter {
	return (*SyncFilter)(New(nbits, nhashes))
}

// Add insert a key with hash value h into f.
//
// The upper and lower half of h are treated as two independent hashes.
// These are used to derive further values using the enhanced double hashing
// construction of Dillinger and Manolios,
// https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf.
func (f *SyncFilter) Add(h uint64) {
	h1, h2 := uint32(h>>32), uint32(h)
	b := (*Filter)(f).getblock(h2)

	for i := 1; i < f.k; i++ {
		h1, h2 = doublehash(h1, h2, i)
		b.setbitAtomic(h1)
	}
}

// Has reports whether a key with hash value h has been added.
// It may return a false positive.
func (f *SyncFilter) Has(h uint64) bool {
	h1, h2 := uint32(h>>32), uint32(h)
	b := (*Filter)(f).getblock(h2)

	for i := 1; i < f.k; i++ {
		h1, h2 = doublehash(h1, h2, i)
		if !b.getbitAtomic(h1) {
			return false
		}
	}
	return true
}

// getbitAtomic reports whether bit (i modulo BlockBits) is set.
func (b *block) getbitAtomic(i uint32) bool {
	bit := uint32(1) << (i % wordSize)
	x := atomic.LoadUint32(&(*b)[(i/wordSize)%blockWords])
	return x&bit != 0
}

// setbit sets bit (i modulo BlockBits) of b, atomically.
func (b *block) setbitAtomic(i uint32) {
	bit := uint32(1) << (i % wordSize)
	p := &(*b)[(i/wordSize)%blockWords]

	for {
		old := atomic.LoadUint32(p)
		if old&bit != 0 {
			// Checking here instead of checking the return value from
			// the CAS is between 50% and 80% faster on the benchmark.
			return
		}
		atomic.CompareAndSwapUint32(p, old, old|bit)
	}
}
