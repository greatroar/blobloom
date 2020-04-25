Blobloom
========

A blocked Bloom filter package for Go (golang) with no runtime dependencies.

[Blocked Bloom filters](https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf)
are a cache-efficient variant of Bloom filters, the well-known approximate set
data structure. To quote [Daniel Lemire](https://lemire.me/blog/2019/12/19/xor-filters-faster-and-smaller-than-bloom-filters/),
they have "unbeatable speed".

Blobloom does not provide hash functions for use with the Bloom filter.
Instead, it requires you, the user, to supply hash values. That means you get
to pick the hash algorithm that is fastest for your data or reuse hashes that
you've already computed (say, a SHA-2). You only need to supply one 64-bit
hash value or two 32-bit ones, as Blobloom uses the [enhanced double
hashing](https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf)
algorithm to synthesize any further hash values it needs.

See the [package documentation](https://godoc.org/github.com/greatroar/blobloom)
for usage information and examples.

Performance
-----------

This package contains benchmarks for comparison against other Bloom filter
packages. To run these benchmarks, pick a build tag from the following table:

| Tag            | Package                                  |
| -------------- | ---------------------------------------- |
| (no tag)       | This package with pre-hashed inputs      |
| bbloom         | github.com/ipfs/bbloom                   |
| blobloomxxhash | This package + github.com/cespare/xxhash |
| willf          | github.com/willf/bloom                   |

Then edit compare_$tag_test.go to remove the line

    // +build ignore

and invoke go test as follows:

    go test -run='^$' -tags "benchcompare $tag" -bench=.

Omit -tags and its argument to run the benchmarks for Blobloom. These assume
that the input keys (which are random strings) can be used as hashes without
any processing. This reflects the original use case (in [Syncthing](
https://syncthing.net)) where SHA-256 hashes where stored in a Bloom filter.
If this does not describe your use case, benchmark with the tag blobloomxxhash
to run the keys through the [xxhash](https://github.com/cespare/xxhash)
function.

Benchmarking a different package changes the go.mod file. The result is not
checked in, because it would make the benchmarked packages a dependency of
Blobloom and transitively of every user.

On an Intel Core i7-3770k (3.5GHz), Blobloom is between two and four times
faster than bbloom on a Bloom filter that can handle 100 million elements with
an FPR of 0.1%:

    Add1e8_1e2-8         166ns ± 1%    66ns ±22%  -60.32%  (p=0.000 n=9+10)
    Add1e8_1e3-8         202ns ± 1%    91ns ± 2%  -55.00%  (p=0.000 n=10+10)
    TestEmpty1e8_1e2-8  59.0ns ± 2%  13.8ns ± 1%  -76.57%  (p=0.000 n=5+10)
    TestEmpty1e8_1e3-8  65.4ns ± 2%  14.3ns ± 3%  -78.11%  (p=0.001 n=5+10)
    TestNeg1e8_1e2-8    64.4ns ± 2%  21.9ns ± 0%  -65.92%  (p=0.002 n=10+4)
    TestPos1e8_1e2-8    68.9ns ± 4%  18.7ns ± 1%  -72.93%  (p=0.008 n=5+5)
    TestPos1e8_1e3-8     106ns ± 8%    25ns ± 4%  -76.88%  (p=0.008 n=5+5)

When we add in xxhash, it's still 40% faster:

    Add1e8_1e2-8         166ns ± 1%    96ns ± 1%  -42.51%  (p=0.000 n=9+10)
    Add1e8_1e3-8         202ns ± 1%   105ns ± 1%  -47.69%  (p=0.000 n=10+10)
    TestNeg1e8_1e2-8    64.4ns ± 2%  37.4ns ± 1%  -41.90%  (p=0.001 n=10+5)
    TestNeg1e8_1e3-8    62.5ns ± 2%  38.6ns ± 2%  -38.26%  (p=0.008 n=5+5)
    TestPos1e8_1e2-8    68.9ns ± 4%  35.1ns ± 1%  -49.12%  (p=0.008 n=5+5)
    TestPos1e8_1e3-8     106ns ± 8%    40ns ± 0%  -62.31%  (p=0.008 n=5+5)

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
