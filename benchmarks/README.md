This module contains benchmarks for comparison against other Bloom filter
packages. To run these benchmarks, pick a build tag from the following table:

| Tag      | Package                                  |
| -------- | ---------------------------------------- |
| (no tag) | This package with pre-hashed inputs      |
| bbloom   | github.com/ipfs/bbloom                   |
| xxhash   | This package + github.com/cespare/xxhash |
| willf    | github.com/willf/bloom                   |

Then invoke go test as follows:

    go test -tags="$tag" -bench=.

Omit -tags and its argument to run the benchmarks for Blobloom. These assume
that the input keys (which are random strings) can be used as hashes without
any processing. This reflects the original use case (in [Syncthing](
https://syncthing.net)) where SHA-256 hashes where stored in a Bloom filter.
If this does not describe your use case, benchmark with the tag blobloomxxhash
to run the keys through the [xxhash](https://github.com/cespare/xxhash)
function.

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
