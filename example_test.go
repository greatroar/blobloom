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
		// one in a million, but we only have 2GiB (= 2^30 bytes) to spare.
		FPRate:  1e-6,
		MaxBits: 8 << 31,
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
