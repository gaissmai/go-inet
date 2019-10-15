package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

var flagPrintTree = flag.Bool("t", false, "print as tree")
var progname string

func main() {
	progname = filepath.Base(os.Args[0])

	log.SetPrefix(progname + ": ")
	log.SetFlags(0)

	// get and check flags
	flag.Usage = usage
	flag.Parse()

	if len(flag.Args()) < 2 || len(flag.Args()) > 3 {
		usage()
	}

	// get and check CIDR
	startCidr, err := inet.ParseBlock(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	if !startCidr.IsCIDR() {
		log.Fatalf("invalid input, no CIDR: %s", startCidr)
	}

	// get and check bits
	bits := [2]int{}
	for i, s := range flag.Args()[1:] {
		bit, err := strconv.Atoi(s)
		if err != nil {
			log.Fatal(err)
		}
		if bit <= 0 || bit > 20 {
			log.Fatal(fmt.Errorf("wrong bits: %d,  must be 0 < bits < 21, wrong bits", bit))
		}
		bits[i] = bit
	}

	// first split
	cidrs := startCidr.SplitCIDR(bits[0])

	// second split
	if bits[1] != 0 {
		for _, c := range cidrs {
			cidrs = append(cidrs, c.SplitCIDR(bits[1])...)
		}
	}

	if *flagPrintTree {
		// make tree.Items from cidrs
		items := make([]tree.Item, len(cidrs))
		for i, c := range cidrs {
			items[i] = tree.Item{c, nil, nil}
		}
		tr := tree.New()
		tr.Insert(items...)
		tr.Fprint(os.Stdout)
	} else {
		inet.SortBlock(cidrs)
		for _, c := range cidrs {
			fmt.Println(c)
		}
	}
}

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage of %s:\n\n", progname)
	fmt.Fprintf(w, "$ %s [-t] start bits [bits]\n\n", progname)
	flag.PrintDefaults()

	example := `
example:

$ cidrsplit -t 2001:db8:900::/48 1 2
▼
├─ 2001:db8:900::/49
│  ├─ 2001:db8:900::/51
│  ├─ 2001:db8:900:2000::/51
│  ├─ 2001:db8:900:4000::/51
│  └─ 2001:db8:900:6000::/51
└─ 2001:db8:900:8000::/49
   ├─ 2001:db8:900:8000::/51
   ├─ 2001:db8:900:a000::/51
   ├─ 2001:db8:900:c000::/51
   └─ 2001:db8:900:e000::/51
`
	fmt.Fprint(w, example)
	os.Exit(1)
}
