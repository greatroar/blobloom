// Copyright 2020-2021 the Blobloom authors
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

// This package implements a toy interactive spell checker.
//
// It reads a dictionary from /usr/share/dict/words and a text from standard
// input. It then reports any misspelled words on standard output.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/maphash"
	"log"
	"os"
	"unicode"

	"github.com/greatroar/blobloom"
)

func main() {
	dict := loadDictionary()

	sc := bufio.NewScanner(os.Stdin)
	sc.Split(bufio.ScanWords)

	for sc.Scan() {
		word := normalize(sc.Bytes())
		if !dict.has(word) {
			fmt.Printf(">>> %s\n", word)
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

// A Bloom filter with a randomized hash function.
type bloomfilter struct {
	*blobloom.Filter
	maphash.Seed
}

func newBloomfilter(capacity uint64, fprate float64) *bloomfilter {
	cfg := blobloom.Config{Capacity: capacity, FPRate: .001}
	return &bloomfilter{
		Filter: blobloom.NewOptimized(cfg),
		Seed:   maphash.MakeSeed(),
	}
}

func (f *bloomfilter) add(key []byte)      { f.Filter.Add(f.hash(key)) }
func (f *bloomfilter) has(key []byte) bool { return f.Filter.Has(f.hash(key)) }

func (f *bloomfilter) hash(key []byte) uint64 {
	var h maphash.Hash
	h.SetSeed(f.Seed)
	h.Write(key)
	return h.Sum64()
}

func normalize(word []byte) []byte {
	word = bytes.TrimFunc(word, unicode.IsPunct)
	word = bytes.ToLower(word)
	return word
}

// To estimate the number of keys without scanning the file twice, we need
// an estimate of the average length of a word. This comes close for English.
const avgWordLength = 10

func loadDictionary() *bloomfilter {
	f, err := os.Open("/usr/share/dict/words")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	filesize := uint64(info.Size())

	dict := newBloomfilter(filesize/avgWordLength, .001)

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		dict.add(normalize(sc.Bytes()))
		filesize-- // Subtract newline, for fairness.
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("dictionary loaded: %dkiB on disk, %dkiB in memory",
		filesize/1024, dict.NumBits()/(8*1024))

	return dict
}
