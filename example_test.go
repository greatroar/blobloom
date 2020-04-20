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

package blobloom_test

import (
	"fmt"
	"hash/fnv"
	"io"

	"github.com/greatroar/blobloom"
)

func Example_fnv() {
	f := blobloom.New(10000, 5)
	h := fnv.New64()

	messages := []string{
		"Hello!",
		"Welcome!",
		"Mind your step!",
		"Have fun!",
		"Goodbye!",
	}

	for _, msg := range messages {
		h.Reset()
		io.WriteString(h, msg)
		f.Add(h.Sum64())
	}

	for _, msg := range messages {
		h.Reset()
		io.WriteString(h, msg)
		if f.Has(h.Sum64()) {
			fmt.Println(msg)
		} else {
			panic("Bloom filter didn't get the message")
		}
	}

	// Output:
	// Hello!
	// Welcome!
	// Mind your step!
	// Have fun!
	// Goodbye!
}

func ExampleOptimize() {
	cfg := blobloom.Config{
		// We want to insert a billion keys and get a false positive rate of
		// one in a million, but we only have 2GiB (= 2^31 bytes) to spare.
		Capacity: 1e9,
		FPRate:   1e-6,
		MaxBits:  8 * 1 << 31,
	}
	nbits, nhashes := blobloom.Optimize(cfg)
	fpr := blobloom.FPRate(cfg.Capacity, nbits, nhashes)

	// How big will the filter be and what FP rate will we achieve?
	fmt.Printf("size = %dMiB\nfpr = %.3f\n", nbits/(8<<20), fpr)

	// Output:
	// size = 2048MiB
	// fpr = 0.001
}

const nworkers = 4

func getKeys(keys chan<- string) {
	keys <- "hello"
	keys <- "goodbye"
	close(keys)
}

func hash(key string) uint64 {
	h := fnv.New64()
	io.WriteString(h, key)
	return h.Sum64()
}

func ExampleFilter_Union() {
	// Union can be used to fill a Bloom filter using multiple goroutines.
	//
	// Each goroutine allocates a filter, so the memory use increases by
	// a factor nworkers-1 compared to a sequential version or a version
	// using AddAtomic.

	keys := make(chan string, nworkers)
	filters := make(chan *blobloom.Filter, nworkers)

	go getKeys(keys)

	for i := 0; i < nworkers; i++ {
		go func() {
			f := blobloom.New(1<<20, 6)
			for key := range keys {
				f.Add(hash(key))
			}

			filters <- f
		}()
	}

	f := <-filters
	for i := 1; i < nworkers; i++ {
		f.Union(<-filters)
	}

	// Output:
}
