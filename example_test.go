// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
		f.Add64(h.Sum64())
	}

	for _, msg := range messages {
		h.Reset()
		io.WriteString(h, msg)
		if f.Has64(h.Sum64()) {
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
		FPRate:  1e-6,
		MaxBits: 8 * 1 << 31,
		NKeys:   1e9,
	}
	nbits, nhashes := blobloom.Optimize(cfg)
	fpr := blobloom.FPRate(cfg.NKeys, nbits, nhashes)

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
				f.Add64(hash(key))
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
