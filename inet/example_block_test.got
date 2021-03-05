package inet_test

import (
	"fmt"
	"net"
	"sort"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleParseBlock() {
	for _, anyOf := range []interface{}{
		"fe80::1-fe80::2",         // block from string
		"10.0.0.0-11.255.255.255", // block from string, as range but true CIDR, see output

		net.IP{192, 168, 0, 0},         // IP from net.IP
		inet.MustIP("2001:db8:900::1"), // IP from inet.IP
	} {
		a, _ := inet.ParseBlock(anyOf)
		fmt.Printf("block: %v\n", a)
	}

	// Output:
	// block: fe80::1-fe80::2
	// block: 10.0.0.0/7
	// block: 192.168.0.0/32
	// block: 2001:db8:900::1/128

}

func ExampleSortBlock() {
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
		buf = append(buf, inet.MustBlock(s))
	}

	sort.Slice(buf, func(i, j int) bool { return buf[i].Less(buf[j]) })
	fmt.Printf("%v\n", buf)

	// Output:
	// [0.0.0.0/0 10.0.0.0/9 10.96.0.2-10.96.1.17 ::/0 ::/112 2001:db8::/32 2001:db8:dea0::/44]

}

func ExampleBlock_FindFreeCIDR_v4() {
	outer := inet.MustBlock("192.168.2.0/24")

	inner := make([]inet.Block, 0)
	inner = append(inner, inet.MustBlock("192.168.2.0/26"))
	inner = append(inner, inet.MustBlock("192.168.2.240-192.168.2.249"))

	free := outer.FindFreeCIDR(inner)
	fmt.Printf("%v - %v\nfree: %v\n", outer, inner, free)

	// Output:
	// 192.168.2.0/24 - [192.168.2.0/26 192.168.2.240-192.168.2.249]
	// free: [192.168.2.64/26 192.168.2.128/26 192.168.2.192/27 192.168.2.224/28 192.168.2.250/31 192.168.2.252/30]

}

func ExampleBlock_FindFreeCIDR_v6() {
	outer := inet.MustBlock("2001:db8:de00::/40")

	inner := make([]inet.Block, 0)
	inner = append(inner, inet.MustBlock("2001:db8:dea0::/44"))

	free := outer.FindFreeCIDR(inner)
	fmt.Printf("%v - %v\nfree: %v\n", outer, inner, free)

	// Output:
	// 2001:db8:de00::/40 - [2001:db8:dea0::/44]
	// free: [2001:db8:de00::/41 2001:db8:de80::/43 2001:db8:deb0::/44 2001:db8:dec0::/42]

}

func ExampleBlock_SplitCIDR_v6() {
	a := inet.MustBlock("2001:db8:dea0::/44")
	splits := a.SplitCIDR(1)
	fmt.Println(a, splits)

	// Output:
	// 2001:db8:dea0::/44 [2001:db8:dea0::/45 2001:db8:dea8::/45]

}

func ExampleBlock_SplitCIDR_v4() {
	a := inet.MustBlock("127.0.0.1/8")
	splits := a.SplitCIDR(2)
	fmt.Println(a, splits)

	// Output:
	// 127.0.0.0/8 [127.0.0.0/10 127.64.0.0/10 127.128.0.0/10 127.192.0.0/10]

}

func ExampleBlock_BlockToCIDRList() {
	for _, s := range []string{
		"10.128.0.0-10.128.2.7",
		"2001:b8::3-2001:b8::f",
	} {
		a := inet.MustBlock(s)
		fmt.Printf("%v -> %v\n", a, a.BlockToCIDRList())
	}

	// Output:
	// 10.128.0.0-10.128.2.7 -> [10.128.0.0/23 10.128.2.0/29]
	// 2001:b8::3-2001:b8::f -> [2001:b8::3/128 2001:b8::4/126 2001:b8::8/125]

}

func ExampleAggregate() {
	var bs []inet.Block
	for _, s := range []string{
		"10.0.0.0/32",
		"10.0.0.1/32",
		"10.0.0.4/30",
		"10.0.0.7-10.0.0.99",
		"fe80::/12",
		"fe80:0000:0000:0000:fe2d:5eff:fef0:fc64/128",
		"fe80::/10",
	} {
		bs = append(bs, inet.MustBlock(s))
	}

	packed := inet.Aggregate(bs)
	fmt.Printf("%v\n", packed)

	// Output:
	// [10.0.0.0/31 10.0.0.4/30 10.0.0.8/29 10.0.0.16/28 10.0.0.32/27 10.0.0.64/27 10.0.0.96/30 fe80::/10]

}
