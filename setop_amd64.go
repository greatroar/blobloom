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

// +build !nounsafe

package blobloom

import (
	"math/bits"
	"unsafe"
)

// Block reinterpreted as array of uint64.
type block64 [BlockBits / 64]uint64

func (f *Filter) intersect(g *Filter) {
	checkBinop(f, g)

	a, b := f.b, g.b
	for len(a) >= 2 {
		p := (*block64)(unsafe.Pointer(&a[0]))
		q := (*block64)(unsafe.Pointer(&b[0]))

		p[0] &= q[0]
		p[1] &= q[1]
		p[2] &= q[2]
		p[3] &= q[3]
		p[4] &= q[4]
		p[5] &= q[5]
		p[6] &= q[6]
		p[7] &= q[7]

		p = (*block64)(unsafe.Pointer(&a[1]))
		q = (*block64)(unsafe.Pointer(&b[1]))

		p[0] &= q[0]
		p[1] &= q[1]
		p[2] &= q[2]
		p[3] &= q[3]
		p[4] &= q[4]
		p[5] &= q[5]
		p[6] &= q[6]
		p[7] &= q[7]

		a, b = a[2:], b[2:]
	}

	if len(a) > 0 {
		p := (*block64)(unsafe.Pointer(&a[0]))
		q := (*block64)(unsafe.Pointer(&b[0]))

		p[0] &= q[0]
		p[1] &= q[1]
		p[2] &= q[2]
		p[3] &= q[3]
		p[4] &= q[4]
		p[5] &= q[5]
		p[6] &= q[6]
		p[7] &= q[7]
	}
}

func (f *Filter) union(g *Filter) {
	checkBinop(f, g)

	a, b := f.b, g.b
	for len(a) >= 2 {
		p := (*block64)(unsafe.Pointer(&a[0]))
		q := (*block64)(unsafe.Pointer(&b[0]))

		p[0] |= q[0]
		p[1] |= q[1]
		p[2] |= q[2]
		p[3] |= q[3]
		p[4] |= q[4]
		p[5] |= q[5]
		p[6] |= q[6]
		p[7] |= q[7]

		p = (*block64)(unsafe.Pointer(&a[1]))
		q = (*block64)(unsafe.Pointer(&b[1]))

		p[0] |= q[0]
		p[1] |= q[1]
		p[2] |= q[2]
		p[3] |= q[3]
		p[4] |= q[4]
		p[5] |= q[5]
		p[6] |= q[6]
		p[7] |= q[7]

		a, b = a[2:], b[2:]
	}

	if len(a) > 0 {
		p := (*block64)(unsafe.Pointer(&a[0]))
		q := (*block64)(unsafe.Pointer(&b[0]))

		p[0] |= q[0]
		p[1] |= q[1]
		p[2] |= q[2]
		p[3] |= q[3]
		p[4] |= q[4]
		p[5] |= q[5]
		p[6] |= q[6]
		p[7] |= q[7]
	}
}

func (b *block) onescount() (n int) {
	p := (*block64)(unsafe.Pointer(&b[0]))

	n += bits.OnesCount64(p[0])
	n += bits.OnesCount64(p[1])
	n += bits.OnesCount64(p[2])
	n += bits.OnesCount64(p[3])
	n += bits.OnesCount64(p[4])
	n += bits.OnesCount64(p[5])
	n += bits.OnesCount64(p[6])
	n += bits.OnesCount64(p[7])

	return
}
