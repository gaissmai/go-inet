package inet

import (
	"net"
	"testing"
)

func TestMustIP(t *testing.T) {
	MustIP(NewIP([]byte{1, 2, 3, 4}))
	MustIP(NewIP([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))
	MustIP(NewIP(net.IP([]byte{1, 2, 3, 4})))
	netIP := net.IP([]byte{1, 2, 3, 4})
	MustIP(NewIP(&netIP))
	_, _ = NewIP(net.IP(nil))
}

func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	MustIP(NewIP([]byte{1, 2, 3, 4, 5}))
}

func TestIP_IsValid(t *testing.T) {
	if IPZero.IsValid() {
		t.Errorf("IPZero.IsValid() returns true, want false")
	}

	ipv4 := MustIP(NewIP("127.0.0.1"))
	if !ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns false, want true")
	}

	ipv6 := MustIP(NewIP("::1"))
	if !ipv6.IsValid() {
		t.Errorf("ipv6.IsValid() returns false, want true")
	}

	// make ipv4 invalid
	ipv4[17] = 0xff
	if ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns true, want false")
	}

	// make ipv6 invalid
	ipv6[3] = 0xff
	if ipv6.IsValid() {
		t.Errorf("ipv6.IsValid() returns true, want false")
	}
}
