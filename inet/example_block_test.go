package inet_test

import (
	"fmt"
	"net"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleNewBlock() {
	for _, anyOf := range []interface{}{
		"fe80::1-fe80::2",         // block from string
		"10.0.0.0-11.255.255.255", // block from string, as range but true CIDR, see output

		net.IP{192, 168, 0, 0},         // IP from net.IP
		inet.MustIP("2001:db8:900::1"), // IP from inet.IP

		net.IPNet{IP: net.IP{127, 0, 0, 0}, Mask: net.IPMask{255, 0, 0, 0}},  // block from net.IPNet
		&net.IPNet{IP: net.IP{127, 0, 0, 0}, Mask: net.IPMask{255, 0, 0, 0}}, // block from *net.IPNet
	} {
		a, _ := inet.NewBlock(anyOf)
		fmt.Printf("block: %v\n", a)
	}

	// Output:
	// block: fe80::1-fe80::2
	// block: 10.0.0.0/7
	// block: 192.168.0.0/32
	// block: 2001:db8:900::1/128
	// block: 127.0.0.0/8
	// block: 127.0.0.0/8

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

	inet.SortBlock(buf)
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

func ExampleBlock_Version() {
	for _, s := range []string{
		"10.0.0.1/8",
		"fe80::/10",
		"::-::1",
	} {
		a, _ := inet.NewBlock(s)
		fmt.Println(a.Version())
	}

	// Output:
	// 4
	// 6
	// 6
}

func ExampleBlock_MarshalText() {
	for _, s := range []string{
		"127.0.0.0/8",
		"fe80::/10",
		"10.0.0.0-10.1.0.0",
		"",
	} {
		a, _ := inet.NewBlock(s)
		bs, _ := a.MarshalText()
		fmt.Printf("%-20v %-20q %v\n", string(bs), string(bs), bs)
	}

	// Output:
	// 127.0.0.0/8          "127.0.0.0/8"        [49 50 55 46 48 46 48 46 48 47 56]
	// fe80::/10            "fe80::/10"          [102 101 56 48 58 58 47 49 48]
	// 10.0.0.0-10.1.0.0    "10.0.0.0-10.1.0.0"  [49 48 46 48 46 48 46 48 45 49 48 46 49 46 48 46 48]
	//                      ""                   []

}

func ExampleBlock_UnmarshalText() {
	var a = new(inet.Block)
	for _, s := range []string{
		"127.0.000.255/8",         // base gets truncated by CIDR mask, see output
		"10.000.000.000-10.1.0.0", // leading zeros are normalized, see output
		"",                        // empty input string aka []byte(nil) returns zero-value (BlockZero) on UnmarshalText()
		"fe80::",                  // invalid
	} {
		err := a.UnmarshalText([]byte(s))
		if err != nil {
			fmt.Println("ERROR:", err)
			continue
		}
		fmt.Printf("%q\n", a)
	}

	// Output:
	// "127.0.0.0/8"
	// "10.0.0.0-10.1.0.0"
	// ""
	// ERROR: invalid Block

}

func ExampleBlock_Compare() {
	a := inet.MustBlock("127.0.0.0/8")
	b := inet.MustBlock("127.0.0.0/8")
	fmt.Printf("Block{%v}.Compare(Block{%v}) = %d\n", a, b, a.Compare(b))

	a = inet.MustBlock("0.0.0.0/0")
	b = inet.MustBlock("::/0")
	fmt.Printf("Block{%v}.Compare(Block{%v}) = %d\n", a, b, a.Compare(b))

	a = inet.MustBlock("127.128.0.0/9")
	b = inet.MustBlock("127.0.0.0/8")
	fmt.Printf("Block{%v}.Compare(Block{%v}) = %d\n", a, b, a.Compare(b))

	a = inet.MustBlock("fe80::/10")
	b = inet.MustBlock("fe80::/12")
	fmt.Printf("Block{%v}.Compare(Block{%v}) = %d\n", a, b, a.Compare(b))

	// Output:
	// Block{127.0.0.0/8}.Compare(Block{127.0.0.0/8}) = 0
	// Block{0.0.0.0/0}.Compare(Block{::/0}) = -1
	// Block{127.128.0.0/9}.Compare(Block{127.0.0.0/8}) = 1
	// Block{fe80::/10}.Compare(Block{fe80::/12}) = -1

}

func ExampleBlock_Size() {
	for _, s := range []string{
		"0.0.0.0/12",
		"::/0",
		"10.0.0.0-10.0.0.43",
		"2001:db8::1-2001:db8::ffec",
	} {
		a := inet.MustBlock(s)
		fmt.Printf("%-30v size min. %d bits\n", a, a.BitLen())

		if a.BitLen() < 21 {
			fmt.Printf("%-30v len %s IPs\n", a, a.Size())
		}
	}

	// Output:
	// 0.0.0.0/12                     size min. 20 bits
	// 0.0.0.0/12                     len 1048576 IPs
	// ::/0                           size min. 128 bits
	// 10.0.0.0-10.0.0.43             size min. 6 bits
	// 10.0.0.0-10.0.0.43             len 44 IPs
	// 2001:db8::1-2001:db8::ffec     size min. 16 bits
	// 2001:db8::1-2001:db8::ffec     len 65516 IPs

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
		"10.0.0.7/31",
		"fe80::/12",
		"fe80:0000:0000:0000:fe2d:5eff:fef0:fc64/128",
		"fe80::/10",
	} {
		bs = append(bs, inet.MustBlock(s))
	}

	packed := inet.Aggregate(bs)
	fmt.Printf("%v\n", packed)

	// Output:
	// [10.0.0.0/31 10.0.0.4/30 fe80::/10]

}
