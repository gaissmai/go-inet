package inet_test

import (
	"fmt"
	"sort"

	"github.com/gaissmai/go-inet/v2/inet"
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
	ip1, _ := inet.ParseIP("192.168.2.1")
	ip2, _ := inet.ParseIP("fffe:db8::")

	fmt.Printf("%q\n", ip1.Expand())
	fmt.Printf("%q\n", ip2.Expand())

	// Output:
	// "192.168.002.001"
	// "fffe:0db8:0000:0000:0000:0000:0000:0000"
}

func ExampleIP_Reverse() {
	ip1, _ := inet.ParseIP("192.168.2.1")
	ip2, _ := inet.ParseIP("fffe:db8::")

	fmt.Printf("%q\n", ip1.Reverse())
	fmt.Printf("%q\n", ip2.Reverse())

	// Output:
	// "1.2.168.192"
	// "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.e.f.f.f"
}

func ExampleIP_Less() {
	s := []string{
		"0.0.0.1",
		"fe80::1",
		"0.0.0.0",
		"127.0.0.1",
		"::",
		"::1",
		"255.255.255.255",
	}
	var ips []inet.IP
	for _, v := range s {
		ip, _ := inet.ParseIP(v)
		ips = append(ips, ip)
	}

	sort.Slice(ips, func(i, j int) bool { return ips[i].Less(ips[j]) })
	for _, ip := range ips {
		fmt.Println(ip)
	}

	// Output:
	// 0.0.0.0
	// 0.0.0.1
	// 127.0.0.1
	// 255.255.255.255
	// ::
	// ::1
	// fe80::1

}
