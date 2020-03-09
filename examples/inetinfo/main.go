// Command inetinfo, print information about IP addresses, CIDRs and ranges.
//
// Usage: inetinfo ip_or_block
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

	// get and check flags and args
	if len(os.Args) != 2 {
		usage()
	}

	arg := os.Args[1]
	if arg == "-h" || arg == "--h" || arg == "--help" {
		usage()
	}

	// parse arg as IP
	ip, errIP := inet.ParseIP(arg)
	if errIP == nil {
		printIPInfo(ip)
		os.Exit(0)
	}

	// parse arg as Block
	block, errBlock := inet.ParseBlock(arg)
	if errBlock == nil {
		printBlockInfo(block)
		os.Exit(0)
	}

	// wrong CIDR?
	i := strings.IndexByte(arg, '/')
	if i >= 0 {
		log.Fatal(errBlock)
	}

	// wrong range?
	i = strings.IndexByte(arg, '-')
	if i >= 0 {
		log.Fatal(errBlock)
	}

	// wrong IP
	log.Fatal(errIP)
}

// ############################################################

func printIPInfo(ip inet.IP) {
	fmt.Printf("%-10s %v\n", "Version:", ip.Version())
	fmt.Printf("%-10s %v\n", "RFC:", ip)
	fmt.Printf("%-10s %v\n", "Expand:", ip.Expand())
	fmt.Printf("%-10s %v\n", "Reverse:", ip.Reverse())
}

func printBlockInfo(block inet.Block) {
	fmt.Printf("%-10s %v\n", "Version:", block.Version())
	if block.IsCIDR() {
		fmt.Printf("%-10s %v\n", "Prefix:", block)
		if block.Version() == 6 {
			fmt.Printf("%-10s %v\n", "Hints:", cidrHints(block))
		}
		fmt.Printf("%-10s %v\n", "Base:", block.Base)
		fmt.Printf("%-10s %v\n", "Last:", block.Last)
		fmt.Printf("%-10s %v\n", "Mask:", block.Mask)
		fmt.Printf("%-10s %v\n", "Wildcard:", hostmask(block.Mask))
		fmt.Printf("%-10s %v bits\n", "Bits:", block.BitLen())
		fmt.Printf("%-10s %v addrs\n", "Size:", block.Size())
	} else {
		fmt.Printf("%-10s %v-%v\n", "Range:", block.Base, block.Last)
		fmt.Printf("%-10s %v bits (min)\n", "Bits:", block.BitLen())
		fmt.Printf("%-10s %v addrs\n", "Size:", block.Size())
		fmt.Printf("%-10s %v\n", "CIDRList:", block.BlockToCIDRList())
	}
}

// helper
func hostmask(netmask inet.IP) inet.IP {
	nm := netmask.Bytes()
	hostmask := make([]byte, len(nm))
	for i := range nm {
		hostmask[i] = ^nm[i]
	}
	return inet.MustIP(hostmask)
}

// helper for CIDR v6 representation hints:
//
// hint for hextet border (just expand):   => expanded-base::/bits e.g. 2001:0db8::/32
// hint for nibble border but NOT hextet:  => expanded-base~~/bits e.g. 2001:0db8:0~~/36
// hint for NOT on nibble border:          => expanded-base<</bits e.g. 2001:0db8:0<</33
//
func cidrHints(cidr inet.Block) string {
	// get the bits from mask
	bits, _ := net.IPMask(cidr.Mask.Bytes()).Size()

	// expand the base IP
	base := cidr.Base.Expand()

	// no hints for /0 and /128
	if bits == 0 {
		return cidr.String()
	}

	if bits == 128 {
		return fmt.Sprintf("%s/%d", base, bits)
	}

	// calc nibbles and colons for bit mask
	nibbles := int(bits / 4)
	colons := int(bits / 16)

	// is the bit mask on a nibble border
	nibbleBorder := bits%4 == 0

	// is the bit mask also on a hextet border
	hextetBorder := bits%16 == 0

	// default for nibbleBorder is '~~'
	hint := "~~"

	// mask is within a nibble, hint is '<<'
	if !nibbleBorder {
		nibbles++
		hint = "<<"
	}

	// mask is at a hextet border, hint is the default '::'
	if hextetBorder {
		colons--
		hint = "::"
	}

	// cut expanded base due to bit mask
	base = base[0 : nibbles+colons]

	return fmt.Sprintf("%s%s/%d", base, hint, bits)
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
Hints:     2001:0db8:000c:0000:0000:0000:0000:0~~/116
Base:      2001:db8:c::
Last:      2001:db8:c::fff
Mask:      ffff:ffff:ffff:ffff:ffff:ffff:ffff:f000
Wildcard:  ::fff
Bits:      12 bits
Size:      4096 addrs
`
	fmt.Fprint(w, output)
	os.Exit(1)
}
