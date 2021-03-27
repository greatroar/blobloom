This module contains benchmarks for comparison against other Bloom filter
packages. To run these benchmarks, pick a build tag from the following table:

| Tag      | Package                                                     |
| -------- | ----------------------------------------------------------- |
| (no tag) | This package with pre-hashed inputs                         |
| bbloom   | github.com/ipfs/bbloom                                      |
| boom     | github.com/tylertreat/BoomFilters ("classic" Bloom filters) |
| sync     | This package's SyncFilter with pre-hashed inputs            |
| willf    | github.com/willf/bloom                                      |
| xxhash   | This package + github.com/cespare/xxhash                    |
| xxh3     | This package + github.com/zeebo/xxh3                        |

Then invoke go test as follows:

    go test -tags="$tag" -bench=.

Omit -tags and its argument to run the benchmarks for Blobloom. These assume
that the input keys (which are random strings) can be used as hashes without
any processing. This reflects the original use case (in [Syncthing](
https://syncthing.net)) where SHA-256 hashes were stored in a Bloom filter.
If this does not describe your use case, benchmark with the tag xxhash
to run the keys through the [xxhash](https://github.com/cespare/xxhash)
function.

The benchmarks are set up to work with the benchstat tool.
To compare Blobloom+xxh3 to bbloom, do

    go get golang.org/x/perf/cmd/benchstat
    go test -bench=. -count=5 -timeout=30m -tags "bbloom" | tee bbloom.bench
    go test -bench=. -count=5 -timeout=30m -tags "xx3"    | tee xxh3.bench
    benchstat bbloom.bench xxh3.bench

The sync benchmark only measures sequential performance.
