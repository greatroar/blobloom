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

package blobloom

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPRate(t *testing.T) {
	t.Parallel()

	// FP rate is zero when no keys have been inserted.
	assert.EqualValues(t, 0, FPRate(0, 100, 3))

	// FP rate is close to one when the capacity is greatly exceeded.
	nhashes := 100.0 * math.Ln2
	assert.InDelta(t, 1.0, FPRate(1e9, 1e8, int(nhashes)), 1e-7)

	// Examples from Putze et al., page 4.

	// XXX We compute 0.023041, which is confirmed by PARI/GP and SciPy.
	// Is the rounding in the paper off?
	assert.InDelta(t, 0.0231, FPRate(1, 8, 5), 6e-5)

	// XXX This one is only accurate to one digit.
	// The required number does not occur in the series expansion either,
	// the closest partial sum being 1.9536e-4.
	assert.InDelta(t, 1.94e-4, FPRate(1, 20, 14), 3e-5)
}

func TestFPRateConvergence(t *testing.T) {
	for _, c := range []struct {
		c, k float64
		iter int
	}{
		{.01, 1, 2500},
		{.1, 1, 2000},
		{3, 2, 200},
		{4, 2, 200},
		{6, 3, 200},
		{8, 5, 200},
		{20, 14, 100},
		{30, 20, 100},
	} {
		t.Run(fmt.Sprintf("c=%f,k=%d", c.c, int(c.k)), func(t *testing.T) {
			t.Parallel()

			fpr, iterations := fpRate(c.c, c.k)
			t.Logf("fpr = %f", fpr)
			assert.Less(t, iterations, c.iter)
		})
	}
}

func TestFPRateCorrectC(t *testing.T) {
	t.Parallel()

	// Try to reconstruct the correction table. We may be one bit off.
	for i, expect := range correctC[1:] {
		c := float64(i + 1)
		k := float64(c) * math.Ln2
		fprBlock := math.Exp(logFprBlock(c, k))

		cprime := c
		for {
			if p, _ := fpRate(cprime, k); p <= fprBlock {
				break
			}
			cprime++
			k = cprime * math.Ln2
		}

		assert.InDelta(t, float64(expect), cprime, 1)
	}
}

func TestFPRateInvalidInput(t *testing.T) {
	assert.Panics(t, func() { FPRate(10, 0, 2) })
	assert.Panics(t, func() { FPRate(10, 2, 0) })
}

func TestNewOptimizedMaxFPR(t *testing.T) {
	t.Parallel()

	f := NewOptimized(Config{
		Capacity: 0,
		FPRate:   1,
	})
	assert.EqualValues(t, BlockBits, f.NumBits())
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
		assert.EqualValues(t, 0, nbits%BlockBits)

		f := New(nbits, nhashes)
		assert.Equal(t, c.expect, f.NumBits())
	}
}

func TestOptimizeFewBits(t *testing.T) {
	t.Parallel()

	for _, config := range []Config{
		{
			Capacity: 1,
			FPRate:   .99,
			MaxBits:  1,
		},
		{
			Capacity: 100000,
			FPRate:   0.01,
			MaxBits:  408,
		},
	} {
		// Optimize should give nbits >= BlockBits.
		nbits, nhashes := Optimize(config)
		assert.EqualValues(t, BlockBits, nbits)
		assert.Greater(t, nhashes, 0)
	}
}

func TestOptimizeInvalidFPRate(t *testing.T) {
	t.Parallel()

	assert.Panics(t, func() { Optimize(Config{FPRate: 0}) })
	assert.Panics(t, func() { Optimize(Config{FPRate: 1.0000001}) })
}
