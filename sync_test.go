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
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSync(t *testing.T) {
	const (
		nbits    = 1 << 13
		nhashes  = 4
		nworkers = 4
	)

	var (
		hashes = make([]uint64, 1e4)
		r      = rand.New(rand.NewSource(0xaeb15))
		ref    = New(nhashes, nbits)
	)

	for i := range hashes {
		h := r.Uint64()
		hashes[i] = h
		ref.Add(h)
	}

	check := func(f *SyncFilter) {
		t.Helper()

		assert.False(t, ref.Empty())
		assert.Equal(t, ref.b, f.b)

		for i := 0; i < 2e4; i++ {
			h := r.Uint64()
			assert.Equal(t, ref.Has(h), f.Has(h))
		}
	}

	t.Run("all hashes", func(t *testing.T) {
		// Each worker adds all hashes to f.
		t.Parallel()

		f := NewSync(nhashes, nbits)

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
			f  = NewSync(nhashes, nbits)
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
