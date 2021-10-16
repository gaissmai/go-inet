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

type record struct {
	b inet.Block
	t string
}

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
	records := readData(os.Stdin)

	// filter by startBlock
	if (startBlock != inet.Block{}) {
		records = filter(records, startBlock)
	}

	// box block and text to inettree.Item, implements tree.Interface
	items := make([]tree.Interface, 0)
	for _, r := range records {
		items = append(items, boxing(r.b, r.t))
	}

	// find free blocks
	if *flagFree {
		items = free(items)
	}

	// build tree
	t, err := tree.New(items)
	if err != nil {
		fmt.Println("ERROR:", err)
		log.Fatalf("duplicate blocks: %v", t.Duplicates())
	}

	// print tree
	fmt.Println(t)
}

// input records as CSV data:
// block, text...
func readData(in io.Reader) []record {
	out := make([]record, 0)

	r := csv.NewReader(os.Stdin)
	r.FieldsPerRecord = -1

	for {
		fields, err := r.Read()
		if err == io.EOF {
			break
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

		var text string
		if len(fields) > 1 {
			text = strings.Join(fields[1:], " ")
			text = strings.TrimSpace(text)
		}

		// save record
		out = append(out, record{b: block, t: text})
	}
	return out
}

// box the inet.Block and text to inettree.Item, implements tree.Interface
func boxing(b inet.Block, t string) inettree.Item {
	asStr := b.String()
	if t != "" {
		// 10.0.0.0/8 ........................................ RFC-1981
		t = asStr + " " + strings.Repeat(".", 50-len(asStr)) + " " + t
	}
	return inettree.Item{Block: b, Text: t}
}

// filters input blocks by startBlock
func filter(in []record, outer inet.Block) []record {
	out := make([]record, 0, len(in))

	for _, r := range in {
		if outer.Covers(r.b) || outer == r.b {
			out = append(out, r)
		}
	}
	return out
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
	walkFn := func(_ int, item, _ tree.Interface, childs []tree.Interface) error {
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

	if err := t.Walk(walkFn); err != nil {
		log.Fatalf("ERROR, in WalkTreeFn: %v", err)
	}

	for _, b := range free {
		is = append(is, boxing(b, "FREE"))
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
	fmt.Fprintln(w, description)
	os.Exit(1)
}
