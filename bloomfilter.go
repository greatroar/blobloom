// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package blobloom implements blocked Bloom filters.
//
// Blocked Bloom filters are an efficient approximate set data structure.
// Compared to standard Bloom filters, they use the CPU cache more efficiently.
// They are described in detail in a 2010 paper by Putze, Sanders and Singler,
// http://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf.
package blobloom

const (
	MaxBits = (1 << 32) * blockBits
)

// A Filter is a blocked Bloom filter.
type Filter struct {
	b []block // Shards.
	k int     // Number of hash functions required.
}

// New constructs a Bloom filter of size nbits that uses nhashes hash functions.
//
// nhashes must be at least two. The client passes the first two hashes for
// every key to Add and Has, which synthesize all following hashes from the
// two values passed in.
func New(nbits, nhashes int) *Filter {
	if nbits < 1 {
		panic("need at least one bit and two hash functions")
	}
	if nhashes < 2 {
		nhashes = 2
	}
	if nbits > MaxBits {
		panic("nbits exceeds MaxBits")
	}

	// Round nbits up to a multiple of blockBits.
	if nbits%blockBits != 0 {
		nbits += blockBits - nbits%blockBits
	}

	return &Filter{
		b: make([]block, nbits/blockBits),
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

func (f *Filter) NBits() int {
	return blockBits * len(f.b)
}

const (
	// Block size in bytes.
	// This is hardcoded to the L1 cache line size of amd64 and arm64.
	blockSize = 64
	blockBits = 8 * blockSize
)

// A block is a fixed-size Bloom filter, used as a shard of a Filter.
type block [blockSize / 8]uint64

// getbit reports whether bit (i modulo blockBits) is set.
func (b *block) getbit(i uint32) bool {
	const n = uint32(len(*b))
	x := (*b)[(i/64)%n] & (1 << (i % 64))
	return x != 0
}

// setbit sets bit (i modulo blockBits) of b.
func (b *block) setbit(i uint32) {
	const n = uint32(len(*b))
	(*b)[(i/64)%n] |= 1 << (i % 64)
}
