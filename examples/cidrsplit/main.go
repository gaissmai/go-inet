package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gaissmai/go-inet/inet"
)

var flagPrintTree = flag.Bool("t", false, "print as tree")

func init() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()
}

func main() {

	if len(flag.Args()) < 2 || len(flag.Args()) > 3 {
		log.Fatal(fmt.Errorf(usage()))
	}

	// get and check CIDR
	startCidr, err := inet.NewBlock(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
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
	firstCidrs := startCidr.SplitCIDR(bits[0])
	secondCidrs := []inet.Block{}
	for _, cidr := range firstCidrs {
		// second split
		cidrs := cidr.SplitCIDR(bits[1])
		secondCidrs = append(secondCidrs, cidrs...)
	}

	if *flagPrintTree {
		tree := inet.NewTree()
		tree.InsertBulk(firstCidrs)
		tree.InsertBulk(secondCidrs)
		tree.Fprint(os.Stdout)
	} else {
		for _, cidr := range firstCidrs {
			fmt.Println(cidr)
		}
		for _, cidr := range secondCidrs {
			fmt.Println(cidr)
		}
	}
}

func usage() string {
	return `Usage: cidrsplit [-t] start bits [bits]

e.g.  $ cidrsplit 2001:db8:900::/48 4 3
`
}
