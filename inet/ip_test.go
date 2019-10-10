package inet

import (
	"net"
	"testing"
)

func TestMustIP(t *testing.T) {
	MustIP([]byte{1, 2, 3, 4})
	MustIP([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	MustIP(net.IP([]byte{1, 2, 3, 4}))
	netIP := net.IP([]byte{1, 2, 3, 4})
	MustIP(&netIP)
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
	if IPZero.IsValid() {
		t.Errorf("IPZero.IsValid() returns true, want false")
	}

	ipv4 := MustIP("127.0.0.1")
	if !ipv4.IsValid() {
		t.Errorf("ipv4.IsValid() returns false, want true")
	}

	ipv6 := MustIP("::1")
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

func TestIP_MarshalUnmarshalIPZero(t *testing.T) {
	text, err := IPZero.MarshalText()
	if err != nil {
		t.Errorf("marshal IPZero has error: %v", err)
	}
	if string(text) != "" {
		t.Errorf("marshal IPZero isn't \"\"")
	}

	got := new(IP)
	err = got.UnmarshalText(text)
	if err != nil {
		t.Errorf("unmarshal []byte has error: %v", err)
	}

	if *got != IPZero {
		t.Errorf("marshal/unmarshal IPZero isn't idempotent")
	}
}
