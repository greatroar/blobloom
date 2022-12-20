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

import (
	"bytes"
	"math"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	const (
		nkeys    = 1e4
		nworkers = 4
	)

	var (
		config = Config{Capacity: nkeys, FPRate: 1e-5}
		hashes = make([]uint64, nkeys)
		r      = rand.New(rand.NewSource(0xaeb15))
		ref    = NewOptimized(config)
	)

	for i := range hashes {
		h := r.Uint64()
		hashes[i] = h
		ref.Add(h)
	}

	card := ref.Cardinality()
	require.False(t, ref.Empty())
	require.False(t, math.IsInf(card, 0))

	check := func(f *SyncFilter) {
		t.Helper()

		assert.Equal(t, ref.b, f.b)
		assert.False(t, f.Empty())

		for i := 0; i < 2e4; i++ {
			h := r.Uint64()
			assert.Equal(t, ref.Has(h), f.Has(h))
		}
		assert.Equal(t, card, f.Cardinality())

		// Write the filter
		var b bytes.Buffer
		err := f.Write(&b)
		assert.Equal(t, err, nil)

		// Read the filter
		f1, err := ReadSync(&b)
		assert.Equal(t, err, nil)
		assert.Equal(t, true, f1.Equals(f))
	}

	t.Run("all hashes", func(t *testing.T) {
		// Each worker adds all hashes to f.
		t.Parallel()

		f := NewSyncOptimized(config)
		assert.True(t, f.Empty())

		var wg sync.WaitGroup
		wg.Add(nworkers)

		for i := 0; i < nworkers; i++ {
			go func() {
				for _, h := range hashes {
					f.Add(h)
				}
				wg.Done()
			}()
		}

		wg.Wait()
		check(f)
	})

	t.Run("split hashes", func(t *testing.T) {
		// Hashes divided across workers.
		t.Parallel()

		var (
			ch = make(chan uint64, nworkers)
			f  = NewSyncOptimized(config)
			wg sync.WaitGroup
		)
		wg.Add(nworkers)

		go func() {
			for _, h := range hashes {
				ch <- h
			}
			close(ch)
		}()

		for i := 0; i < nworkers; i++ {
			go func() {
				for h := range ch {
					f.Add(h)
				}
				wg.Done()
			}()
		}

		wg.Wait()
		check(f)
	})
}
