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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPRate(t *testing.T) {
	// Examples from Putze et al.
	// XXX The approximation isn't very precise.
	assert.InDeltaf(t, 0.0231, FPRate(1, 8, 6), 3e-4, "")
	assert.InDeltaf(t, 0.000194, FPRate(1, 20, 14), 3e-5, "")
}

func TestNewOptimizedMaxFPR(t *testing.T) {
	f := NewOptimized(Config{
		FPRate: 1,
		NKeys:  0,
	})
	assert.Equal(t, uint64(BlockBits), f.NBits())
}

func TestMaxBits(t *testing.T) {
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
			FPRate:  1e-10,
			NKeys:   2 * c.want,
			MaxBits: c.want,
		})
		// Optimize should round down to multiple of BlockBits.
		assert.LessOrEqual(t, nbits, c.expect)
		assert.Equal(t, uint64(0), nbits%BlockBits)

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
	assert.Equal(t, uint64(BlockBits), f.NBits())
	assert.Equal(t, 2, f.k)
}
