// Bloomstat is a utility for estimating Bloom filter sizes.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/greatroar/blobloom"
)

const usage = `usage: bloomstat capacity false-positive-rate [max-memory]
	The maximum memory may be specified as "10MB", "1.5GiB", etc.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	var (
		capacity = parse("capacity", os.Args[1])
		fpr      = parse("false positive rate", os.Args[2])
		maxsize  float64
	)
	if len(os.Args) > 3 {
		maxsize = parseMem(os.Args[3])
	}

	bits, hashes := blobloom.Optimize(blobloom.Config{
		Capacity: uint64(capacity),
		FPRate:   fpr,
		MaxBits:  uint64(8 * maxsize),
	})

	size, unit := memsize(float64(bits))
	bitsPerKey := float64(bits) / capacity

	expectedFpr := blobloom.FPRate(uint64(capacity), bits, hashes)

	fmt.Printf("%d bits, %.02f %s\n"+
		"%.02f bits/%.02f B per key\n"+
		"%d hashes\n"+
		"%.04f expected false positive rate\n",
		bits, size, unit, bitsPerKey, bitsPerKey/8, hashes, expectedFpr)
}

const (
	kiB = 1 << 10
	MiB = 1 << 20
	GiB = 1 << 30
)

func memsize(bits float64) (size float64, unit string) {
	size = float64(bits) / 8

	switch {
	case size >= GiB:
		size /= GiB
		unit = "GiB"
	case size >= MiB:
		size /= MiB
		unit = "MiB"
	case size >= kiB:
		size /= kiB
		unit = "kiB"
	default:
		unit = "B"
	}
	return
}

func parse(name, num string) float64 {
	v, err := strconv.ParseFloat(num, 64)

	switch e := err.(type) {
	case nil:
	case *strconv.NumError:
		log.Fatalf("%s %q: %v", name, e.Num, e.Err)
	default:
		log.Fatalf("%s: %v", name, err)
	}
	if v < 0 {
		log.Fatalf("%s must be >= 0", name)
	}

	return v
}

func parseMem(s string) float64 {
	var (
		size float64
		unit string
	)
	_, err := fmt.Sscanf(s, "%f%s", &size, &unit)
	if err != nil {
		log.Fatal("max memory:", err)
	}

	switch strings.ToLower(unit) {
	case "kb", "kib":
		size *= kiB
	case "mb", "mib":
		size *= MiB
	case "gb", "gib":
		size *= GiB
	}

	return size
}
