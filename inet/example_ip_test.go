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

func ExampleIP_Compare() {
	a := inet.MustIP("127.0.0.1")
	b := inet.MustIP("10.10.0.1")
	fmt.Printf("IP{%v}.Compare(IP{%v}) = %d\n", a, b, a.Compare(b))

	a = inet.MustIP("0.0.0.0")
	b = inet.MustIP("::")
	fmt.Printf("IP{%v}.Compare(IP{%v}) = %d\n", a, b, a.Compare(b))

	a = inet.MustIP("fe80::1")
	b = inet.MustIP("fe80::1")
	fmt.Printf("IP{%v}.Compare(IP{%v}) = %d\n", a, b, a.Compare(b))

	// Output:
	// IP{127.0.0.1}.Compare(IP{10.10.0.1}) = 1
	// IP{0.0.0.0}.Compare(IP{::}) = -1
	// IP{fe80::1}.Compare(IP{fe80::1}) = 0

}

func ExampleIP_String() {
	for _, ip := range []inet.IP{
		inet.MustIP("127.0.0.1"),
		inet.MustIP("fe80::1"),
		inet.IP(""),
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
		inet.IP(""),
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

func ExampleIP_AddUint64() {
	for _, ip := range []inet.IP{
		inet.MustIP("127.0.0.1"),
		inet.MustIP("2001:db8::"),
	} {
		z := ip.AddUint64(1)
		fmt.Printf("%-10v + %d = %v\n", ip, 1, z)
	}

	// Output:
	// 127.0.0.1  + 1 = 127.0.0.2
	// 2001:db8:: + 1 = 2001:db8::1

}

func ExampleIP_AddBytes() {
	for _, tt := range []struct {
		ip  string
		add []byte
	}{
		{"0.0.0.0", []byte{255, 255, 255, 255}},
		{"::", []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	} {
		ip := inet.MustIP(tt.ip)
		z := ip.AddBytes(tt.add)
		fmt.Printf("%v + %#x = %v\n", tt.ip, tt.add, z)
	}

	// Output:
	// 0.0.0.0 + 0xffffffff = 255.255.255.255
	// :: + 0xffffffffffffffffffffffffffffffff = ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff

}

func ExampleIP_SubUint64() {
	for _, ip := range []inet.IP{
		inet.MustIP("127.0.0.0"),
		inet.MustIP("2001:db8::"),
	} {
		z := ip.SubUint64(1)
		fmt.Printf("%-10v - %#x = %v\n", ip, 1, z)
	}

	// Output:
	// 127.0.0.0  - 0x1 = 126.255.255.255
	// 2001:db8:: - 0x1 = 2001:db7:ffff:ffff:ffff:ffff:ffff:ffff

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

func ExampleIP_Bytes() {
	for _, ip := range []inet.IP{
		inet.MustIP("192.168.2.1"),
		inet.MustIP("::ffff:192.168.2.1"), // IP4-mapped
		inet.MustIP("2001:db8:dead::beef"),
	} {
		buf := ip.Bytes()
		fmt.Printf("%#v\n", buf)
	}

	// Output:
	// []byte{0xc0, 0xa8, 0x2, 0x1}
	// []byte{0xc0, 0xa8, 0x2, 0x1}
	// []byte{0x20, 0x1, 0xd, 0xb8, 0xde, 0xad, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xbe, 0xef}

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
	} {
		fmt.Println(ip.Version())
	}

	// Output:
	// 4
	// 4
	// 6
	// 6
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
