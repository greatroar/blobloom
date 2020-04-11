// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package blobloom implements blocked Bloom filters.
//
// Blocked Bloom filters are approximate set data structures: if a key has
// been added to the filter, a lookup of that returns true, but if the key
// has not been added, there is a non-zero chance that the lookup still
// returns true (a false positive).
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
// (two in a million), its uses 20% more memory. At 1e-10, the space required
// is double that of standard Bloom filter.
//
// For more details, see the 2010 paper by Putze, Sanders and Singler,
// https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf.
package blobloom

// BlockBits is the number of bits per block and the minimum number of bits
// in a Filter.
//
// The value of this constant is chosen to match the L1 cache line size
// of popular architectures (386, amd64, arm64).
const BlockBits = 512

// MaxBits is the maximum number of bits supported by a Filter (256GiB).
const MaxBits = BlockBits << 32

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
func New(nbits, nhashes int) *Filter {
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
// construction described by Kirsch and Mitzenmacher,
// https://www.eecs.harvard.edu/~michaelm/postscripts/rsa2008.pdf.
func (f *Filter) Add(h1, h2 uint32) {
	_ = f.b[0] // Suppress divide by zero check.

	i := h1 % uint32(len(f.b))
	b := &f.b[i]

	// Derive k hash functions from h1 and h2
	// using the construction described by Kirsch and Mitzenmacher.
	h := h1
	for i := 1; i < f.k; i++ {
		h += h2
		b.setbit(h)
	}
}

// Add64 calls Add with the upper/lower 32 bits of h as h1/h2.
func (f *Filter) Add64(h uint64) {
	f.Add(uint32(h>>32), uint32(h))
}

// Has reports whether a key with hash values h1 and h2 has been added.
// It may return a false positive.
func (f *Filter) Has(h1, h2 uint32) bool {
	_ = f.b[0] // Suppress divide by zero check.

	i := h1 % uint32(len(f.b))
	b := &f.b[i]

	h := h1
	for i := 1; i < f.k; i++ {
		h += h2
		if !b.getbit(h) {
			return false
		}
	}
	return true
}

// Has64 calls Has with the upper/lower 32 bits of h as h1/h2.
func (f *Filter) Has64(h uint64) bool {
	return f.Has(uint32(h>>32), uint32(h))
}

// NBits returns the number of bits of f.
func (f *Filter) NBits() int {
	return BlockBits * len(f.b)
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

// setbit sets bit (i modulo BlockBits) of b.
func (b *block) setbit(i uint32) {
	const n = uint32(len(*b))
	(*b)[(i/64)%n] |= 1 << (i % 64)
}
