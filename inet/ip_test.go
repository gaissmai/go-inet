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

func TestIP_IsValid(t *testing.T) {
	ipv4 := MustIP("127.0.0.1")
	if !ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns false, want true")
	}

	ipv6mappedipv4 := MustIP("::ffff:127.31.3.2")
	if !ipv6mappedipv4.IsValid() {
		t.Errorf("ipv6mappedipv4.IsValid() returns false, want true")
	}
	if ipv6mappedipv4.Version() != 4 {
		t.Errorf("ipv6mappedipv4.Version() returns %d, want 4", ipv6mappedipv4.Version())
	}

	ipv6 := MustIP("::1")
	if !ipv6.IsValid() {
		t.Errorf("ipv6.IsValid() returns false, want true")
	}

	// make ipv4 invalid
	ipv4 = IP([]byte{4, 127, 0, 0, 1, 0, 0, 0, 0xff})
	if ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns true, want false")
	}

	// make ipv4 invalid
	ipv4 = IP([]byte{5, 127, 0, 0, 1})
	if ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns true, want false")
	}
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
