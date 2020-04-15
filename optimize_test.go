// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPRate(t *testing.T) {
	// Examples from Putze et al.
	// XXX The approximation isn't very precise.
	assert.InDeltaf(t, 0.0231, FPRate(1, 8, 6), 5e-4, "")
	assert.InDeltaf(t, 0.000194, FPRate(1, 20, 14), 3e-5, "")
}

func TestNewOptimizedMaxFPR(t *testing.T) {
	f := NewOptimized(Config{
		FPRate: 1,
		NKeys:  0,
	})
	assert.Equal(t, BlockBits, f.NBits())
}

func TestMaxBits(t *testing.T) {
	for _, c := range []struct {
		want, expect int
	}{
		{1, BlockBits},
		{BlockBits - 1, BlockBits},
		{BlockBits + 1, BlockBits},
		{2*BlockBits - 1, BlockBits},
		{4<<20 - 1, 4<<20 - BlockBits},
		{4<<20 + 1, 4 << 20},
		{4<<20 + BlockBits, 4<<20 + BlockBits},
	} {
		nbits, nhashes := Optimize(Config{
			// Ask for tiny FPR with a huge number of keys.
			FPRate:  1e-10,
			NKeys:   2 * c.want,
			MaxBits: c.want,
		})
		// Optimize should round down to multiple of BlockBits.
		assert.LessOrEqual(t, nbits, c.expect)
		assert.Equal(t, 0, nbits%BlockBits)

		// New should correct cases < BlockBits.
		f := New(nbits, nhashes)
		assert.Equal(t, c.expect, f.NBits())
	}
}

func TestOptimizeOneBitOneHash(t *testing.T) {
	// This configuration produces one hash function.
	nbits, nhashes := Optimize(Config{
		FPRate:  .99,
		MaxBits: 1,
		NKeys:   1,
	})
	assert.Equal(t, 1, nhashes)

	f := New(nbits, nhashes)
	assert.Equal(t, BlockBits, f.NBits())
	assert.Equal(t, 2, f.k)
}
