// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package blobloom implements blocked Bloom filters.
//
// Blocked Bloom filters are an approximate set data structure: if a key has
// been added to a filter, a lookup of that key returns true, but if the key
// has not been added, there is a non-zero probability that the lookup still
// returns true (a false positive). It follows that, if the lookup for a key
// returns false, that key has not been added to the filter.
//
// In this package, keys are represented exclusively as hashes. Client code
// is responsible for supplying two 32-bit hash values for a key. No hash
// function is provided, since the "right" hash function for an application
// depends on the data the application processes.
//
// Compared to standard Bloom filters, blocked Bloom filters use the CPU
// cache more efficiently. A blocked Bloom filter is an array of ordinary
// Bloom filters of fixed size BlockBits (the blocks). The first hash of a
// key selects the block to use.
//
// To achieve the same false positive rate (FPR) as a standard Bloom filter,
// a blocked Bloom filter requires more memory. For an FPR of at most 2e-6
// (two in a million), it uses ~20% more memory. At 1e-10, the space required
// is double that of standard Bloom filter.
//
// For more details, see the 2010 paper by Putze, Sanders and Singler,
// https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf.
package blobloom

import "sync/atomic"

// BlockBits is the number of bits per block and the minimum number of bits
// in a Filter.
//
// The value of this constant is chosen to match the L1 cache line size
// of popular architectures (386, amd64, arm64).
const BlockBits = 512

// MaxBits is the maximum number of bits supported by a Filter.
const MaxBits = BlockBits << 32 // 256GiB.

// A Filter is a blocked Bloom filter.
type Filter struct {
	b []block // Shards.
	k int     // Number of hash functions required.
}

// New constructs a Bloom filter with given numbers of bits and hash functions.
//
// The number of bits should be at least BlockBits; smaller values are silently
// increased.
//
// The number of hash functions uses is silently increased to two.
// The client passes the first two hashes for every key to Add and Has,
// which synthesize all following hashes from the two values passed in.
func New(nbits uint64, nhashes int) *Filter {
	if nbits < 1 {
		nbits = BlockBits
	}
	if nhashes < 2 {
		nhashes = 2
	}
	if nbits > MaxBits {
		panic("nbits exceeds MaxBits")
	}

	// Round nbits up to a multiple of BlockBits.
	if nbits%BlockBits != 0 {
		nbits += BlockBits - nbits%BlockBits
	}

	return &Filter{
		b: make([]block, nbits/BlockBits),
		k: nhashes,
	}
}

// Add inserts a key with hash values h1 and h2 into f.
//
// The two hash values supplied are used to derive further values using the
// enhanced double hashing construction of Dillinger and Manolios,
// https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf.
func (f *Filter) Add(h1, h2 uint32) {
	_ = f.b[0] // Suppress divide by zero check.

	i := h1 % uint32(len(f.b))
	b := &f.b[i]

	// Derive k hash functions from h1 and h2
	// using the construction described by Kirsch and Mitzenmacher.
	for i := 0; i+1 < f.k; i++ {
		h1, h2 = doublehash(h1, h2, i)
		b.setbit(h1)
	}
}

// Add64 calls Add with the upper/lower 32 bits of h as h1/h2.
func (f *Filter) Add64(h uint64) {
	f.Add(uint32(h>>32), uint32(h))
}

// AddAtomic atomically inserts a key with hash values h1 and h2 into f.
//
// Multiple goroutines may call AddAtomic and AddAtomic64 concurrently,
// though no goroutines should call any other methods on f concurrently
// with these methods.
func (f *Filter) AddAtomic(h1, h2 uint32) {
	_ = f.b[0] // Suppress divide by zero check.

	i := h1 % uint32(len(f.b))
	b := &f.b[i]

	for i := 0; i+1 < f.k; i++ {
		h1, h2 = doublehash(h1, h2, i)
		b.setbitAtomic(h1)
	}
}

// AddAtomic64 calls AddAtomic with the upper/lower 32 bits of h as h1/h2.
func (f *Filter) AddAtomic64(h uint64) {
	f.AddAtomic(uint32(h>>32), uint32(h))
}

// Clear resets f to its empty state.
func (f *Filter) Clear() {
	for i := range f.b {
		f.b[i] = block{}
	}
}

// Has reports whether a key with hash values h1 and h2 has been added.
// It may return a false positive.
func (f *Filter) Has(h1, h2 uint32) bool {
	_ = f.b[0] // Suppress divide by zero check.

	i := h1 % uint32(len(f.b))
	b := &f.b[i]

	for i := 0; i+1 < f.k; i++ {
		h1, h2 = doublehash(h1, h2, i)
		if !b.getbit(h1) {
			return false
		}
	}
	return true
}

// doublehash generates the hash values n1, n2 to use in iteration i of
// enhanced double hashing from the values h1, h2 of the previous iteration.
func doublehash(h1, h2 uint32, i int) (uint32, uint32) {
	h1 = h1 + h2
	h2 = h2 + uint32(i)
	return h1, h2
}

// Has64 calls Has with the upper/lower 32 bits of h as h1/h2.
func (f *Filter) Has64(h uint64) bool {
	return f.Has(uint32(h>>32), uint32(h))
}

// NBits returns the number of bits of f.
func (f *Filter) NBits() uint64 {
	return BlockBits * uint64(len(f.b))
}

// Union sets f to the union of f and g.
//
// Union panics when f and g do not have the same number of bits and
// hash functions. Both Filters must be using the same hash function(s),
// but Union cannot check this.
func (f *Filter) Union(g *Filter) {
	if len(f.b) != len(g.b) {
		panic("Bloom filters do not have the same number of bits")
	}
	if f.k != g.k {
		panic("Bloom filters do not have the same number of hash functions")
	}
	for i := range f.b {
		f.b[i].union(&g.b[i])
	}
}

const blockSize = BlockBits / 64

// A block is a fixed-size Bloom filter, used as a shard of a Filter.
type block [blockSize]uint64

// getbit reports whether bit (i modulo BlockBits) is set.
func (b *block) getbit(i uint32) bool {
	const n = uint32(len(*b))
	x := (*b)[(i/64)%n] & (1 << (i % 64))
	return x != 0
}

func (b *block) union(c *block) {
	b[0] |= c[0]
	b[1] |= c[1]
	b[2] |= c[2]
	b[3] |= c[3]
	b[4] |= c[4]
	b[5] |= c[5]
	b[6] |= c[6]
	b[7] |= c[7]
}

// setbit sets bit (i modulo BlockBits) of b.
func (b *block) setbit(i uint32) {
	const n = uint32(len(*b))
	(*b)[(i/64)%n] |= 1 << (i % 64)
}

// setbit sets bit (i modulo BlockBits) of b, atomically.
func (b *block) setbitAtomic(i uint32) {
	const n = uint32(len(*b))
	bit := uint64(1) << (i % 64)
	p := &(*b)[(i/64)%n]
	for {
		old := atomic.LoadUint64(p)
		if old&bit != 0 {
			// Checking here instead of checking the return value from
			// the CAS is between 25% and 50% faster on the benchmark.
			return
		}
		new := old | bit
		atomic.CompareAndSwapUint64(p, old, new)
	}
}
