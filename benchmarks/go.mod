module github.com/greatroar/blobloom/benchmarks

go 1.14

require (
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/greatroar/blobloom v0.2.0
	github.com/ipfs/bbloom v0.0.4
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
)

replace github.com/greatroar/blobloom => ../
