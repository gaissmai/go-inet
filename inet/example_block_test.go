package inet_test

import (
	"fmt"
	"net"
	"sort"

	"github.com/gaissmai/go-inet/v2/inet"
)

func ExampleParseBlock() {
	for _, anyOf := range []interface{}{
		"fe80::1-fe80::2",         // block from string
		"10.0.0.0-11.255.255.255", // block from string, as range but true CIDR, see output
		net.IP{192, 168, 0, 0},    // IP from net.IP
	} {
		a, _ := inet.ParseBlock(anyOf)
		fmt.Printf("block: %v\n", a)
	}

	// Output:
	// block: fe80::1-fe80::2
	// block: 10.0.0.0/7
	// block: 192.168.0.0/32

}

func ExampleBlock_Less() {
	var buf []inet.Block
	for _, s := range []string{
		"2001:db8:dead:beef::/44",
		"10.0.0.0/9",
		"::/0",
		"10.96.0.2-10.96.1.17",
		"0.0.0.0/0",
		"::-::ffff",
		"2001:db8::/32",
	} {
		b, _ := inet.ParseBlock(s)
		buf = append(buf, b)
	}

	sort.Slice(buf, func(i, j int) bool { return buf[i].Less(buf[j]) })
	fmt.Printf("%v\n", buf)

	// Output:
	// [0.0.0.0/0 10.0.0.0/9 10.96.0.2-10.96.1.17 ::/0 ::/112 2001:db8::/32 2001:db8:dea0::/44]

}

func ExampleMerge() {
	var bs []inet.Block
	for _, s := range []string{
		"10.0.0.0/32",
		"10.0.0.1/32",
		"10.0.0.4/30",
		"10.0.0.6-10.0.0.99",
		"fe80::/12",
		"fe80:0000:0000:0000:fe2d:5eff:fef0:fc64/128",
		"fe80::/10",
	} {
		b, _ := inet.ParseBlock(s)
		bs = append(bs, b)
	}

	packed := inet.Merge(bs)
	fmt.Printf("%v\n", packed)

	// Output:
	// [10.0.0.0/31 10.0.0.4-10.0.0.99 fe80::/10]

}

func ExampleBlock_CIDRs() {
	b, _ := inet.ParseBlock("10.0.0.6-10.0.0.99")
	fmt.Printf("%v\n", b.CIDRs())

	b, _ = inet.ParseBlock("2001:db8::affe-2001:db8::ffff")
	fmt.Printf("%v\n", b.CIDRs())

	// Output:
	// [10.0.0.6/31 10.0.0.8/29 10.0.0.16/28 10.0.0.32/27 10.0.0.64/27 10.0.0.96/30]
	// [2001:db8::affe/127 2001:db8::b000/116 2001:db8::c000/114]
}

func ExampleBlock_Diff_v4() {
	outer, _ := inet.ParseBlock("192.168.2.0/24")

	inner := make([]inet.Block, 0)
	for _, s := range []string{
		"192.168.2.0/26",
		"192.168.2.240-192.168.2.249",
	} {
		b, _ := inet.ParseBlock(s)
		inner = append(inner, b)
	}

	var got []inet.Block
	for _, b := range outer.Diff(inner) {
		got = append(got, b.CIDRs()...)
	}
	fmt.Printf("%v - %v\ndiff: %v\n", outer, inner, got)

	// Output:
	// 192.168.2.0/24 - [192.168.2.0/26 192.168.2.240-192.168.2.249]
	// diff: [192.168.2.64/26 192.168.2.128/26 192.168.2.192/27 192.168.2.224/28 192.168.2.250/31 192.168.2.252/30]

}

func ExampleBlock_Diff_v6() {

	outer, _ := inet.ParseBlock("2001:db8:de00::/40")
	b, _ := inet.ParseBlock("2001:db8:dea0::/44")
	inner := []inet.Block{b}

	var got []inet.Block
	for _, b := range outer.Diff(inner) {
		got = append(got, b.CIDRs()...)
	}
	fmt.Printf("%v - %v\ndiff: %v\n", outer, inner, got)

	// Output:
	// 2001:db8:de00::/40 - [2001:db8:dea0::/44]
	// diff: [2001:db8:de00::/41 2001:db8:de80::/43 2001:db8:deb0::/44 2001:db8:dec0::/42]

}
