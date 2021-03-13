Blobloom
========

A blocked Bloom filter package for Go (golang) with no runtime dependencies.

[Blocked Bloom filters](https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf)
are a cache-efficient variant of Bloom filters, the well-known approximate set
data structure. To quote [Daniel Lemire](https://lemire.me/blog/2019/12/19/xor-filters-faster-and-smaller-than-bloom-filters/),
they have unbeatable speed. See the directory ``benchmarks/`` to determine
exactly how fast Blobloom is compared to other packages.

Usage
-----

Construct a Bloom filter:

	f := blobloom.NewOptimized(blobloom.Config{
		Capacity: nkeys, // Expected number of keys.
		FPRate:   1e-4,  // One in 10000 false positives is acceptable.
	})

Add a key:

	// import "github.com/cespare/xxhash/v2"
	h := xxhash.Sum64String(key)
	f.Add(h)

Test for presence of a key:

	h := xxhash.Sum64String(key)
	if f.Has(h) {
		// Key is probably in f.
	} else {
		// Key is certainly not present in f.
	}

See the [package documentation](https://pkg.go.dev/github.com/greatroar/blobloom)
for further usage information and examples.

Hash functions
--------------

Blobloom does not provide hash functions. Instead, it requires client code to
represent each key as a single 64-bit hash value, leaving it to the user to
pick the right hash function for a particular problem. Here are some general
suggestions:

* If you use Bloom filters to speed up access to a key-value store, you might
want to look at [xxh3](https://github.com/zeebo/xxh3) or [xxhash](
https://github.com/cespare/xxhash).
* If your keys are cryptographic hashes, consider using the first 8 bytes of those hashes.
* If you use Bloom filters to make probabilistic decisions, a randomized hash
function such as [siphash](https://github.com/dchest/siphash) or [maphash](
https://golang.org/pkg/hash/maphash) may prevent the same false positives
occurring every time.

When evaluating a hash function, or designing a custom one, be aware that
Blobloom uses the [fastrange](
https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/)
reduction on the lower 32 bits of the hash to select a block, and
[enhanced double hashing](https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf)
on the upper and lower 32-bit halves, followed by modulo 2<sup>9</sup>,
to derive bit indices.
In particular, this means that casting a 32-bit hash to uint64 causes the
Bloom filter to perform suboptimally.
(These are details of the current implementation, not API guarantees.)


License
-------

Copyright © 2020-2021 the Blobloom authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
