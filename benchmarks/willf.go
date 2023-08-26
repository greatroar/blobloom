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

//go:build willf
// +build willf

package benchmarks

import "github.com/bits-and-blooms/bloom/v3"

type bloomFilter bloom.BloomFilter

func (f *bloomFilter) Add(hash []byte) {
	((*bloom.BloomFilter)(f)).Add(hash)
}

func (f *bloomFilter) Has(hash []byte) bool {
	return ((*bloom.BloomFilter)(f)).Test(hash)
}

func newBF(capacity int, fpr float64) *bloomFilter {
	f := bloom.NewWithEstimates(uint(capacity), fpr)
	return (*bloomFilter)(f)
}
