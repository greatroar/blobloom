// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build !benchcompare

package blobloom_test

import (
	"encoding/binary"

	"github.com/greatroar/blobloom"
)

type bloomFilter blobloom.Filter

func (f *bloomFilter) Add(hash []byte) {
	h := binary.BigEndian.Uint64(hash[:8])
	((*blobloom.Filter)(f)).Add64(h)
}

func (f *bloomFilter) Has(hash []byte) bool {
	h := binary.BigEndian.Uint64(hash[:8])
	return ((*blobloom.Filter)(f)).Has64(h)
}

func newBF(capacity int, fpr float64) *bloomFilter {
	f := blobloom.NewOptimized(blobloom.Config{
		FPRate: fpr,
		NKeys:  uint64(capacity),
	})
	return (*bloomFilter)(f)
}
