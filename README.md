Blobloom
========

A blocked Bloom filter package for Go (golang) with no external dependencies.

[Blocked Bloom filters](https://algo2.iti.kit.edu/documents/cacheefficientbloomfilters-jea.pdf)
are a cache-efficient variant of Bloom filters, the well-known approximate set
data structure.

Blobloom does not provide hash functions for use with the Bloom filter.
Instead, it requires you, the user, to supply hash values. That means you get
to pick the hash algorithm that is fastest for your data or reuse hashes that
you've already computed (say, a SHA-2). You only need to supply one 64-bit
hash value or two 32-bit ones, as Blobloom uses the [enhanced double
hashing](https://www.ccs.neu.edu/home/pete/pub/bloom-filters-verification.pdf)
algorithm to synthesize any further hash values it needs.

See the [package documentation](https://godoc.org/github.com/greatroar/blobloom)
for usage information and examples.

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
