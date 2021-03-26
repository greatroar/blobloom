module github.com/greatroar/blobloom/benchmarks

go 1.14

require (
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/greatroar/blobloom v0.6.0
	github.com/ipfs/bbloom v0.0.4
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tylertreat/BoomFilters v0.0.0-20210315201527-1a82519a3e43
	github.com/willf/bitset v1.1.11 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	github.com/zeebo/xxh3 v0.10.0
	golang.org/x/sys v0.0.0-20210324051608-47abb6519492 // indirect
)

replace github.com/greatroar/blobloom => ../
