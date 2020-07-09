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
		if !dict.Has(hash(word)) {
			fmt.Printf(">>> %s\n", word)
		}
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}
}

// FNV1a hash function.
func hash(key []byte) uint64 {
	var h uint64 = 0xcbf29ce484222325
	for _, c := range key {
		h ^= uint64(c)
		h *= 0x100000001b3
	}
	return h
}

func normalize(word []byte) []byte {
	word = bytes.ToLower(word)
	word = bytes.TrimFunc(word, unicode.IsPunct)
	return word
}

// To estimate the number of keys without scanning the file twice, we need
// an estimate of the average length of a word. This comes close for English.
const avgWordLength = 10

func loadDictionary() *blobloom.Filter {
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

	dict := blobloom.NewOptimized(blobloom.Config{
		Capacity: filesize / avgWordLength,
		FPRate:   .001,
	})

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		dict.Add(hash(normalize(sc.Bytes())))
		filesize-- // Subtract newline, for fairness.
	}
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	log.Printf("dictionary loaded: %dkiB on disk, %dkiB in memory",
		filesize/1024, dict.NumBits()/(8*1024))

	return dict
}
