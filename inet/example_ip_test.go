package inet_test

import (
	"fmt"
	"net"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleParseIP() {
	for _, anyOf := range []interface{}{
		"2001:db8::",                       // string
		net.IP{10, 0, 0, 1},                // net.IP
		[]byte{127, 0, 0, 1},               // []byte, IPv4
		[]byte{0: 0xff, 1: 0xfe, 15: 0x01}, // []byte, IPv6
	} {
		ip, _ := inet.ParseIP(anyOf)
		fmt.Printf("ip: %v\n", ip)
	}

	// Output:
	// ip: 2001:db8::
	// ip: 10.0.0.1
	// ip: 127.0.0.1
	// ip: fffe::1

}

func ExampleIP_String() {
	for _, ip := range []inet.IP{
		inet.MustIP("127.0.0.1"),
		inet.MustIP("fe80::1"),
		inet.IPZero,
	} {
		fmt.Printf("%#v\n", ip.String())
	}

	// Output:
	// "127.0.0.1"
	// "fe80::1"
	// ""
}

func ExampleIP_MarshalText() {
	for _, ip := range []inet.IP{
		inet.MustIP("127.0.0.1"),
		inet.MustIP("fe80::1"),
		inet.IPZero,
	} {

		bs, err := ip.MarshalText()
		if err != nil {
			fmt.Printf("%.78s ...\n", err)
			continue
		}
		fmt.Printf("%-15v %-15q %v\n", string(bs), string(bs), bs)
	}

	// Output:
	// 127.0.0.1       "127.0.0.1"     [49 50 55 46 48 46 48 46 49]
	// fe80::1         "fe80::1"       [102 101 56 48 58 58 49]
	//                 ""              []
}

func ExampleSortIP() {
	var buf []inet.IP
	for _, s := range []string{
		"2001:db8::",
		"127.0.0.1",
		"::1",
		"::FFFF:0.0.0.1",
	} {
		buf = append(buf, inet.MustIP(s))
	}

	inet.SortIP(buf)
	fmt.Printf("%v\n", buf)

	// Output:
	// [0.0.0.1 127.0.0.1 ::1 2001:db8::]
}

func ExampleIP_Expand() {
	for _, ip := range []inet.IP{
		inet.MustIP("192.168.2.1"),
		inet.MustIP("fffe:db8::"),
	} {

		fmt.Printf("%q\n", ip.Expand())
	}

	// Output:
	// "192.168.002.001"
	// "fffe:0db8:0000:0000:0000:0000:0000:0000"
}

func ExampleIP_Reverse() {
	for _, ip := range []inet.IP{
		inet.MustIP("192.168.2.1"),
		inet.MustIP("fffe:db8::"),
	} {

		fmt.Printf("%q\n", ip.Reverse())
	}

	// Output:
	// "1.2.168.192"
	// "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.e.f.f.f"
}

func ExampleIP_ToNetIP() {
	for _, ip := range []inet.IP{
		inet.MustIP("192.168.2.1"),
		inet.MustIP("fffe:db8::"),
		// {5}, Panics on invalid input.
	} {
		fmt.Printf("%#v\n", ip.ToNetIP())
	}

	// Output:
	// net.IP{0xc0, 0xa8, 0x2, 0x1}
	// net.IP{0xff, 0xfe, 0xd, 0xb8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
}

func ExampleIP_Version() {
	for _, ip := range []inet.IP{
		inet.MustIP("192.168.2.1"),
		inet.MustIP("::ffff:192.168.2.1"), // IP4-mapped
		inet.MustIP("fe80::1"),
		inet.MustIP("2001:db8:dead::beef"),
		inet.IPZero,
	} {
		fmt.Println(ip.Version())
	}

	// Output:
	// 4
	// 4
	// 6
	// 6
	// 0
}

func ExampleIP_UnmarshalText() {
	var ip = new(inet.IP)
	for _, s := range []string{
		"127.0.0.1",
		"fe80::1",
		"", // empty input string aka []byte(nil) returns zero-value (IP{}) on UnmarshalText()
		"ge80::1",
	} {
		err := ip.UnmarshalText([]byte(s))
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%-15q %03d\n", ip, []byte(*ip))
	}

	// Output:
	// "127.0.0.1"     [004 127 000 000 001]
	// "fe80::1"       [006 254 128 000 000 000 000 000 000 000 000 000 000 000 000 000 001]
	// ""              []
	// invalid IP

}
