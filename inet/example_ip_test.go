package inet_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleParseIP() {
	for _, s := range []string{
		"2001:db8::",         // IPv6
		"::ffff:192.168.0.1", // IPv4-mapped IPv6
		"10.0.0.1",           // IPv4
	} {
		ip, _ := inet.ParseIP(s)
		fmt.Printf("ip: %v\n", ip)
	}

	// Output:
	// ip: 2001:db8::
	// ip: 192.168.0.1
	// ip: 10.0.0.1

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
