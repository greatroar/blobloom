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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPRate(t *testing.T) {
	t.Parallel()

	// FP rate is zero when no keys have been inserted.
	assert.Equal(t, 0.0, FPRate(0, 100, 3))

	// Examples from Putze et al., page 4.

	// XXX We compute 0.023041, which is confirmed by PARI/GP and SciPy.
	// Is the rounding in the paper off?
	assert.InDelta(t, 0.0231, FPRate(1, 8, 5), 1e-4)

	// XXX This one is only accurate to one digit.
	// The required number does not occur in the series expansion either,
	// the closest partial sum being 1.9536e-4.
	assert.InDelta(t, 1.94e-4, FPRate(1, 20, 14), 3e-5)
}

func TestFPRateCorrectC(t *testing.T) {
	t.Parallel()

	// Try to reconstruct the correction table. We may be one bit off.
	for i, expect := range correctC[1:] {
		c := float64(i + 1)
		k := float64(c) * math.Ln2
		fprBlock := math.Exp(logFprBlock(c, k))

		cprime := c
		for fpRate(cprime, k) > fprBlock {
			cprime++
			k = cprime * math.Ln2
		}

		assert.InDelta(t, float64(expect), cprime, 1)
	}
}

func TestNewOptimizedMaxFPR(t *testing.T) {
	t.Parallel()

	f := NewOptimized(Config{
		Capacity: 0,
		FPRate:   1,
	})
	assert.Equal(t, uint64(BlockBits), f.NumBits())
}

func TestMaxBits(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		want, expect uint64
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
			Capacity: 2 * c.want,
			FPRate:   1e-10,
			MaxBits:  c.want,
		})
		// Optimize should round down to multiple of BlockBits.
		assert.LessOrEqual(t, nbits, c.expect)
		assert.Equal(t, uint64(0), nbits%BlockBits)

		// New should correct cases < BlockBits.
		f := New(nbits, nhashes)
		assert.Equal(t, c.expect, f.NumBits())
	}
}

func TestOptimizeOneBitOneHash(t *testing.T) {
	t.Parallel()

	// This configuration produces one hash function.
	nbits, nhashes := Optimize(Config{
		Capacity: 1,
		FPRate:   .99,
		MaxBits:  1,
	})
	assert.Equal(t, 1, nhashes)

	// New fixes that up to two, because we need one hash function
	// to select a block.
	f := New(nbits, nhashes)
	assert.Equal(t, uint64(BlockBits), f.NumBits())
	assert.Equal(t, 2, f.k)
}
