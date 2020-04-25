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

// To run the Blobloom benchmarks on ipfs/bbloom, remove the "build ignore"
// line below, then
//
//     go test -run='^$' -tags "benchcompare bbloom" -bench=.
//
// The ignore constraint is there to prevent ipfs/bbloom from ending up in
// go.mod and becoming a transitive dependency for all users.

// +build bbloom

package benchmarks

import "github.com/ipfs/bbloom"

type bloomFilter = bbloom.Bloom

func newBF(capacity int, fpr float64) *bloomFilter {
	f, err := bbloom.New(float64(capacity), fpr)
	if err != nil {
		panic(err)
	}
	return (*bloomFilter)(f)
}
