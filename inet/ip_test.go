package inet

import (
	"net"
	"testing"
)

func TestMustIP(t *testing.T) {
	MustIP("127.0.0.1")
	MustIP("0.0.0.0")
	MustIP("fffe::1")
	FromStdIP(net.IP([]byte{1, 2, 3, 4}))
}

func TestPanic_1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	MustIP("1.2.3.4.5")
}

func TestIP_addOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{MustIP("0.0.0.0"), MustIP("0.0.0.1"), true},
		{MustIP("127.0.0.1"), MustIP("127.0.0.2"), true},
		{MustIP("255.255.255.255"), ipZero, false},
		//
		{MustIP("::"), MustIP("::1"), true},
		{MustIP("::1"), MustIP("::2"), true},
		{MustIP("::ffff:ffff:ffff:ffff"), MustIP("0:0:0:1::"), true},
		{MustIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), ipZero, false},
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
		{MustIP("0.0.0.0"), ipZero, false},
		{MustIP("0.0.0.1"), MustIP("0.0.0.0"), true},
		{MustIP("127.0.0.1"), MustIP("127.0.0.0"), true},
		//
		{MustIP("::"), ipZero, false},
		{MustIP("::1"), MustIP("::"), true},
		{MustIP("0:0:0:1::"), MustIP("::ffff:ffff:ffff:ffff"), true},
	}

	for _, tt := range ips {
		if got, ok := tt.in.subOne(); got != tt.want || ok != tt.ok {
			t.Errorf("(%v).subOne = %v, %v; want: %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
