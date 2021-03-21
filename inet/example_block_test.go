package inet_test

import (
	"fmt"
	"sort"

	"github.com/gaissmai/go-inet/v2/inet"
)

func mustParseBlock(s string) inet.Block {
	b, err := inet.ParseBlock(s)
	if err != nil {
		panic(err)
	}
	return b
}

func ExampleParseBlock() {
	for _, s := range []string{
		"fe80::1-fe80::2",         // block from string
		"10.0.0.0-11.255.255.255", // block from string, as range but true CIDR, see output
	} {
		a, _ := inet.ParseBlock(s)
		fmt.Printf("block: %v\n", a)
	}

	// Output:
	// block: fe80::1-fe80::2
	// block: 10.0.0.0/7
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
	inner := []inet.Block{
		mustParseBlock("192.168.2.0/26"),
		mustParseBlock("192.168.2.240-192.168.2.249"),
	}

	fmt.Printf("%v - %v\ndiff: %v\n", outer, inner, outer.Diff(inner))

	// Output:
	// 192.168.2.0/24 - [192.168.2.0/26 192.168.2.240-192.168.2.249]
	// diff: [192.168.2.64-192.168.2.239 192.168.2.250-192.168.2.255]
}

func ExampleBlock_Diff_v6() {

	outer, _ := inet.ParseBlock("2001:db8:de00::/40")
	inner, _ := inet.ParseBlock("2001:db8:dea0::/44")

	fmt.Printf("%v - %v\ndiff: %v\n", outer, inner, outer.Diff([]inet.Block{inner}))

	// Output:
	// 2001:db8:de00::/40 - 2001:db8:dea0::/44
	// diff: [2001:db8:de00::-2001:db8:de9f:ffff:ffff:ffff:ffff:ffff 2001:db8:deb0::-2001:db8:deff:ffff:ffff:ffff:ffff:ffff]
}
