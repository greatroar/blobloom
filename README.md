Blobloom
========

A blocked Bloom filter package for Go (golang) with no runtime dependencies.

[Blocked Bloom filters](https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf)
are a cache-efficient variant of Bloom filters, the well-known approximate set
data structure. To quote [Daniel Lemire](https://lemire.me/blog/2019/12/19/xor-filters-faster-and-smaller-than-bloom-filters/),
they have unbeatable speed. See the directory ``benchmarks/`` to determine
exactly how fast Blobloom is compared to other packages.

Blobloom does not provide hash functions for use with the Bloom filter.
Instead, it requires client code to supply hash values. That means you get to
pick the hash algorithm that is fastest for your data, use a secure hash such
as SipHash or reuse hashes that you've already computed.
You only need to supply one 64-bit hash value as Blobloom uses the
[enhanced double hashing](https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf)
algorithm to synthesize any further hash values it needs.

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

See the [package documentation](https://godoc.org/github.com/greatroar/blobloom)
for further usage information and examples.

Hash functions
--------------

If you need a fast hash function, try [xxhash](https://github.com/cespare/xxhash).
If you need a secure hash function, look at [siphash](https://github.com/dchest/siphash).
If you're using Go 1.14, [maphash](https://golang.org/pkg/hash/maphash/)
may suit your needs.

License
-------

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
