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

// +build boom

package benchmarks

import "github.com/tylertreat/BoomFilters"

type bloomFilter boom.BloomFilter

func (f *bloomFilter) Add(hash []byte) {
	((*boom.BloomFilter)(f)).Add(hash)
}

func (f *bloomFilter) Has(hash []byte) bool {
	return ((*boom.BloomFilter)(f)).Test(hash)
}

func newBF(capacity int, fpr float64) *bloomFilter {
	f := boom.NewBloomFilter(uint(capacity), fpr)
	f.SetHash(&nopHash{})
	return (*bloomFilter)(f)
}

// No-op hash function. Assumes all data is written to it in one Write call.
type nopHash struct{ data []byte }

func (h *nopHash) BlockSize() int      { return 1 }
func (h *nopHash) Reset()              {}
func (h *nopHash) Size() int           { return 8 }
func (h *nopHash) Sum(d []byte) []byte { return append(d, h.data...) }
func (h *nopHash) Sum64() uint64       { panic("not used by BoomFilters") }

func (h *nopHash) Write(p []byte) (n int, err error) {
	if len(p) > 8 {
		p = p[:8]
	}
	h.data = p
	return len(p), nil
}
