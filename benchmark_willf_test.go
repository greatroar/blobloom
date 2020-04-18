// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build benchcompare willf

package blobloom_test

import "github.com/willf/bloom"

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
