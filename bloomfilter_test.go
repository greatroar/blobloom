// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package blobloom

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	r := rand.New(rand.NewSource(0x758e326))
	keys := make([]uint64, 10000)
	for i := range keys {
		keys[i] = r.Uint64()
	}

	for _, config := range []struct {
		nbits, nhashes int
	}{
		{1, 2},
		{1024, 4},
		{100, 3},
		{10000, 7},
		{1000000, 14},
	} {
		f := New(config.nbits, config.nhashes)
		assert.GreaterOrEqual(t, f.NBits(), config.nbits)
		assert.LessOrEqual(t, f.NBits(), config.nbits+512)

		for _, k := range keys {
			assert.False(t, f.Has64(k))
		}
		for _, k := range keys {
			f.Add64(k)
		}
		for _, k := range keys {
			assert.True(t, f.Has64(k))
		}
	}
}
