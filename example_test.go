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

package blobloom_test

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"sync"

	"github.com/greatroar/blobloom"
)

func Example_fnv() {
	// This example uses the hash/fnv package from the standard Go library.

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

func Example_sha224() {
	// If you have items addressed by a cryptographic hash,
	// you can use a prefix of it as the hash value for a Bloom filter.
	//
	// If the cryptohashes denote objects from an untrusted source,
	// the Bloom filter can be tricked into giving false positives for
	// chosen objects, because it only uses a small part of the hash
	// that can easily be broken (by a birthday attack). If that can
	// cause problems in your application, first run SipHash on the
	// full cryptohash to get the hash value for the Bloom filter:
	//
	//	import "github.com/dchest/siphash"
	//	h := siphash.Hash(secret1, secret2, key[:])

	// A list of files, identified by their SHA-224.
	files := []string{
		"\x85\x52\xd8\xb7\xa7\xdc\x54\x76\xcb\x9e\x25\xde\xe6\x9a\x80\x91\x29\x07\x64\xb7\xf2\xa6\x4f\xe6\xe7\x8e\x95\x68",
		"\xa0\xad\x8f\x63\x90\x72\x74\x7b\xc3\x43\x09\x45\x94\x0e\x7c\x73\xb8\x34\x93\xf1\x77\x90\x0f\xd2\x7d\x09\x65\x94",
		"\x7b\xd3\xdb\x48\x1e\x7b\x05\x2c\x88\x18\x68\xcc\x13\xc3\x04\x34\x43\x2d\x7b\x49\x24\x74\x70\x33\xd2\xe8\x6e\x73",
	}

	// first64 extracts the first 64 bits of a key as a uint64.
	// The choice of big vs. little-endian is arbitrary.
	first64 := func(key []byte) uint64 {
		return binary.BigEndian.Uint64(key[:8])
	}

	f := blobloom.NewOptimized(blobloom.Config{Capacity: 600, FPRate: .002})

	for _, filehash := range files {
		f.Add(first64([]byte(filehash)))
	}

	for _, s := range []string{"Hello, world!", "Goodbye"} {
		h := sha256.Sum224([]byte(s))
		found := f.Has(first64(h[:]))
		if found {
			fmt.Printf("Found: %v\n", s)
		}
	}

	// Output:
	// Found: Hello, world!
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

var hashes [200]uint64

func init() {
	for i := range hashes {
		hashes[i] = uint64(i)
	}
}

func ExampleFilter_Cardinality_infinity() {
	// To handle the case of Cardinality returning +Inf, track the number of
	// calls to Add and compute the minimum.

	// This Bloom filter is constructed with too many hash functions
	// to force +Inf.
	f := blobloom.New(512, 100)
	var numAdded int

	add := func(h uint64) {
		f.Add(h)
		numAdded++
	}

	for _, h := range hashes {
		add(h)
	}

	estimate := f.Cardinality()
	fmt.Printf("blobloom's estimate:    %.2f\n", estimate)
	fmt.Printf("number of calls to Add: %d\n", numAdded)
	estimate = math.Min(estimate, float64(numAdded))
	fmt.Printf("combined estimate:      %.2f\n", estimate)

	// Output:
	// blobloom's estimate:    +Inf
	// number of calls to Add: 200
	// combined estimate:      200.00
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
	// Each goroutine allocates a filter, so the memory use increases
	// by a factor nworkers-1 compared to a sequential version
	// or a SyncFilter.

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

func ExampleSyncFilter() {
	// Multiple goroutines can Add to a SyncFilter concurrently,
	// without requiring separate synchronization.

	f := blobloom.NewSync(1<<20, 6)
	var wg sync.WaitGroup

	add := func(hs []uint64) {
		for _, h := range hs {
			f.Add(h)
		}
		wg.Done()
	}

	wg.Add(2)
	half := len(hashes) / 2
	go add(hashes[:half])
	go add(hashes[half:])

	wg.Wait() // Wait for updating goroutines to complete.

	for _, h := range hashes {
		if !f.Has(h) {
			fmt.Printf("hash %d added but not retrieved\n", h)
		}
	}

	// Output:
}
