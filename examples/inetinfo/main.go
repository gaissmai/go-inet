package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gaissmai/go-inet/inet"
)

var progname string

func main() {
	progname = filepath.Base(os.Args[0])

	log.SetPrefix(progname + ": ")
	log.SetFlags(0)

	if len(os.Args) != 2 {
		usage()
	}

	arg := os.Args[1]
	if arg == "-h" || arg == "--h" || arg == "--help" {
		usage()
	}

	ip, errIP := inet.NewIP(arg)
	if errIP == nil {
		printIPInfo(ip)
		os.Exit(0)
	}

	block, errBlock := inet.NewBlock(arg)
	if errBlock == nil {
		printBlockInfo(block)
		os.Exit(0)
	}

	i := strings.IndexByte(arg, '/')
	if i >= 0 {
		log.Fatal(errBlock)
	}

	i = strings.IndexByte(arg, '-')
	if i >= 0 {
		log.Fatal(errBlock)
	}

	log.Fatal(errIP)
}

func printIPInfo(ip inet.IP) {
	fmt.Printf("%-10s %v\n", "Version:", ip.Version())
	fmt.Printf("%-10s %v\n", "RFC:", ip)
	fmt.Printf("%-10s %v\n", "Expand:", ip.Expand())
	fmt.Printf("%-10s %v\n", "Reverse:", ip.Reverse())
}

func printBlockInfo(block inet.Block) {
	fmt.Printf("%-10s %v\n", "Version:", block.Version())
	if block.Mask != inet.IPZero {
		fmt.Printf("%-10s %v\n", "Prefix:", block)
		fmt.Printf("%-10s %v\n", "Mask:", block.Mask)
		fmt.Printf("%-10s %v\n", "Wildcard:", hostmask(block.Mask))
		fmt.Printf("%-10s %v bits\n", "Bits:", block.BitLen())
		fmt.Printf("%-10s %v addrs\n", "Size:", block.Size())
	} else {
		fmt.Printf("%-10s %v-%v\n", "Range:", block.Base, block.Last)
		fmt.Printf("%-10s %v bits (min)\n", "Bits:", block.BitLen())
		fmt.Printf("%-10s %v addrs\n", "Size:", block.Size())
	}
}

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(w, "Usage of %s:\n\n", progname)
	fmt.Fprintf(w, "$ %s ip_or_block\n\n", progname)
	fmt.Fprint(w, "example:\n")
	fmt.Fprintf(w, "$ %s %s\n", progname, "2001:db8:c::/116")

	output := `
Version:   6
Prefix:    2001:db8:c::/116
Mask:      ffff:ffff:ffff:ffff:ffff:ffff:ffff:f000
Wildcard:  ::fff
Bits:      12 bits
Size:      4096 addrs
`
	fmt.Fprint(w, output)
	os.Exit(1)
}

func hostmask(netmask inet.IP) inet.IP {
	nm := netmask.Bytes()
	hostmask := make([]byte, len(nm))
	for i := range nm {
		hostmask[i] = ^nm[i]
	}
	return inet.MustIP(hostmask)
}
