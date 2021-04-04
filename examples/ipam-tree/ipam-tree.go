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
With the flag -f, free blocks are marked as FREE and also printed.

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

	bs := make([]tree.Interface, 0)
	for b, t := range input {
		bs = append(bs, marshalItem(b, t))
	}

	// filter and find free
	if (startBlock != inet.Block{}) {
		bs = filter(bs, startBlock)
	}
	if *flagFree {
		bs = free(bs)
	}

	// build tree
	t, err := tree.New(bs)
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

// marshal the text field
func marshalItem(b inet.Block, t string) inettree.Item {
	bs := b.String()
	if t != "" {
		t = bs + " " + strings.Repeat(".", 50-len(bs)) + " " + t
	}
	return inettree.Item{Block: b, Text: t}
}

// filters input blocks by startBlock
func filter(bs []tree.Interface, outer inet.Block) []tree.Interface {
	result := make([]tree.Interface, 0, len(bs))

	for _, item := range bs {
		if outer.Covers(item.(inettree.Item).Block) || outer == item.(inettree.Item).Block {
			result = append(result, item)
		}
	}
	return result
}

func assertBlock(is []tree.Interface) (bs []inet.Block) {
	for _, v := range is {
		bs = append(bs, v.(inettree.Item).Block)
	}
	return
}

// find free
func free(is []tree.Interface) []tree.Interface {

	// make tree with input
	t, err := tree.New(is)
	if err != nil {
		fmt.Println("ERROR:", err)
		log.Fatalf("duplicate blocks: %v", t.Duplicates())
	}

	// find free blocks
	var free []inet.Block
	fn := func(_ int, item, _ tree.Interface, childs []tree.Interface) error {
		if childs == nil {
			return nil
		}

		{
			// type assertions from tree.Interface to inet.Block
			item := item.(inettree.Item).Block
			childs := assertBlock(childs)

			// calc free blocks for every item
			for _, diff := range item.Diff(childs) {
				free = append(free, diff.CIDRs()...)
			}
		}
		return nil
	}

	if err := t.Walk(fn); err != nil {
		log.Fatalf("ERROR, in WalkTreeFn: %v", err)
	}

	for _, b := range free {
		is = append(is, marshalItem(b, "FREE"))
	}

	return is
}

// check flags and arguments
func checkCmdline() {
	flag.Usage = usage
	flag.Parse()
	w := flag.CommandLine.Output()

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
