// Command ipam-tree, read blocks from Stdin, print sorted Tree
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/inettree"
	"github.com/gaissmai/go-inet/v2/tree"
)

var flagFree = flag.Bool("f", false, "show also free blocks under startBlock")

var startBlock inet.Block

var description = `
Read records with blocks and text (separated by comma) from STDIN and prints the tree representation.
If a startBlock is defined as argument, the tree is restricted to blocks covered by startBlock.
With the flag -f, free blocks direct under the startBlock are marked as FREE and also printed.

Input:
10.0.0.0/8, RFC-1918
10.0.0.1/24, my home network
2001:db8::/32, documentation only
fd02:b25f:2cb0::/48, my ULA
::1, home sweet home
127.0.0.1, home sweet home

Output:
▼
├─ 10.0.0.0/8 .................... RFC-1918
│  └─ 10.0.0.0/24 ................... my home network
├─ 127.0.0.1/32 .................. home sweet home
├─ ::1/128 ....................... home sweet home
├─ 2001:db8::/32 ................. documentation only
└─ fd02:b25f:2cb0::/48 ........... my ULA
`

func main() {
	checkCmdline()

	// input records
	input := make(map[inet.Block]string)
	readData(input)

	// filter and find free
	if (startBlock != inet.Block{}) {
		if *flagFree {
			input = free(input, startBlock)
		} else {
			input = filter(input, startBlock)
		}
	}

	// augment blocks with text
	bs := make([]tree.Interface, 0)
	for b, t := range input {

		// augment the block string with text from stdin:
		// ::0, home sweet home
		//
		// Output:
		// └─ ::/128 ........................ home sweet home

		if t != "" {
			s := b.String()
			t = s + " " + strings.Repeat(".", 30-len(s)) + " " + t
		}

		bs = append(bs, inettree.Item{b, t})
	}

	// build tree
	t, err := tree.NewTree(bs)
	if err != nil {
		log.Fatal(err)
	}

	// print tree
	fmt.Println(t)
}

// input records as CSV data:
// block, text...
func readData(m map[inet.Block]string) {
	r := csv.NewReader(os.Stdin)
	r.FieldsPerRecord = -1

	for {
		fields, err := r.Read()
		if err == io.EOF {
			return
		}

		if err != nil {
			log.Printf("skip line: %v", err)
			continue
		}

		f0 := strings.TrimSpace(fields[0])

		block, err := inet.ParseBlock(f0)
		if err != nil {
			log.Printf("skip record: %v (%v)", err, f0)
			continue
		}
		if _, ok := m[block]; ok {
			log.Printf("duplicate block: %v", block)
			continue
		}

		var text string
		if len(fields) > 1 {
			text = strings.Join(fields[1:], " ")
			text = strings.TrimSpace(text)
		}

		// save record
		m[block] = text
	}
}

// filters input blocks by startBlock
func filter(input map[inet.Block]string, outer inet.Block) map[inet.Block]string {
	result := make(map[inet.Block]string, len(input))
	for block, text := range input {
		if outer.Covers(block) || outer == block {
			result[block] = text
		}
	}
	return result
}

// filters input blocks by startBlock and calc free blocks, return both
func free(input map[inet.Block]string, outer inet.Block) map[inet.Block]string {

	result := make(map[inet.Block]string, len(input))
	result[outer] = input[outer]

	// filter and copy to result
	inner := make([]inet.Block, 0, len(input))
	for block, text := range input {
		if outer.Covers(block) {
			inner = append(inner, block)
			result[block] = text
		}
	}

	// calc free blocks, copy to result
	for _, freeBlock := range outer.Diff(inner) {
		for _, cidr := range freeBlock.CIDRs() {
			result[cidr] = "FREE"
		}
	}
	return result
}

// check flags and arguments
func checkCmdline() {
	flag.Usage = usage
	flag.Parse()
	w := flag.CommandLine.Output()

	if *flagFree && len(flag.Args()) == 0 {
		fmt.Fprintf(w, "ERROR: missing start block\n\n")
		usage()
	}

	if len(flag.Args()) > 0 {
		block, err := inet.ParseBlock(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(w, "ERROR: wrong start block '%s': %v\n\n", flag.Arg(0), err)
			usage()
		}
		startBlock = block
	}
}

// just the usage
func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage: %s [-f] [startBlock]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(w, description)
	os.Exit(1)
}
