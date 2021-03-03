package inet

import (
	"net"
	"testing"
)

func TestMustIP(t *testing.T) {
	MustIP([]byte{1, 2, 3, 4})
	MustIP([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	MustIP(net.IP([]byte{1, 2, 3, 4}))
	_, _ = ParseIP(net.IP(nil))
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	MustIP([]byte{1, 2, 3, 4, 5})
}

func TestIP_addOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{MustIP("0.0.0.0"), MustIP("0.0.0.1"), true},
		{MustIP("127.0.0.1"), MustIP("127.0.0.2"), true},
		{MustIP("255.255.255.255"), IPZero, false},
		//
		{MustIP("::"), MustIP("::1"), true},
		{MustIP("::1"), MustIP("::2"), true},
		{MustIP("::ffff:ffff:ffff:ffff"), MustIP("0:0:0:1::"), true},
		{MustIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), IPZero, false},
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
		{MustIP("0.0.0.0"), IPZero, false},
		{MustIP("0.0.0.1"), MustIP("0.0.0.0"), true},
		{MustIP("127.0.0.1"), MustIP("127.0.0.0"), true},
		//
		{MustIP("::"), IPZero, false},
		{MustIP("::1"), MustIP("::"), true},
		{MustIP("0:0:0:1::"), MustIP("::ffff:ffff:ffff:ffff"), true},
	}

	for _, tt := range ips {
		if got, ok := tt.in.subOne(); got != tt.want || ok != tt.ok {
			t.Errorf("(%v).subOne = %v, %v; want: %v, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}
