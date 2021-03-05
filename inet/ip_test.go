package inet

import (
	"net"
	"testing"
)

func mustIP(i interface{}) IP {
	ip, err := ParseIP(i)
	if err != nil {
		panic(err)
	}
	return ip
}

func TestParseIP(t *testing.T) {
	mustIP("127.0.0.1")
	mustIP("0.0.0.0")
	mustIP("fffe::1")
	mustIP(net.IP([]byte{1, 2, 3, 4}))
}

func TestPanic_1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	mustIP("1.2.3.4.5")
}

func TestIP_addOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{mustIP("0.0.0.0"), mustIP("0.0.0.1"), true},
		{mustIP("127.0.0.1"), mustIP("127.0.0.2"), true},
		{mustIP("255.255.255.255"), ipZero, false},
		//
		{mustIP("::"), mustIP("::1"), true},
		{mustIP("::1"), mustIP("::2"), true},
		{mustIP("::ffff:ffff:ffff:ffff"), mustIP("0:0:0:1::"), true},
		{mustIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), ipZero, false},
	}

	for _, tt := range ips {
		if got, ok := tt.in.addOne(); got != tt.want || ok != tt.ok {
			t.Errorf("(%v).addOne = %v, %v; want: %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}

}

func TestIP_subOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{mustIP("0.0.0.0"), ipZero, false},
		{mustIP("0.0.0.1"), mustIP("0.0.0.0"), true},
		{mustIP("127.0.0.1"), mustIP("127.0.0.0"), true},
		//
		{mustIP("::"), ipZero, false},
		{mustIP("::1"), mustIP("::"), true},
		{mustIP("0:0:0:1::"), mustIP("::ffff:ffff:ffff:ffff"), true},
	}

	for _, tt := range ips {
		if got, ok := tt.in.subOne(); got != tt.want || ok != tt.ok {
			t.Errorf("(%v).subOne = %v, %v; want: %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
