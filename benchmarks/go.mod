module github.com/greatroar/blobloom/benchmarks

go 1.14

require (
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/d4l3k/messagediff v1.2.1 // indirect
	github.com/greatroar/blobloom v0.2.0
	github.com/ipfs/bbloom v0.0.4
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/tylertreat/BoomFilters v0.0.0-20200520150052-42a7b4300c0c
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	github.com/zeebo/xxh3 v0.8.2
)

replace github.com/greatroar/blobloom => ../
