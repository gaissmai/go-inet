package inet

import (
	"net"
	"testing"
)

func mustIP(i interface{}) IP {
	switch v := i.(type) {
	case string:
		if ip, err := ParseIP(v); err == nil {
			return ip
		}
		panic(errInvalidIP)
	case net.IP:
		if ip, err := FromStdIP(v); err == nil {
			return ip
		}
		panic(errInvalidIP)
	default:
		panic(errInvalidIP)
	}
}

func TestIPZero(t *testing.T) {
	var ip IP

	if got := ip.String(); got != errInvalidIP.Error() {
		t.Errorf("String() for ipZero; expected %s, got %v", errInvalidIP, got)
	}

	if got := ip.Expand(); got != errInvalidIP.Error() {
		t.Errorf("Expand() for ipZero; expected %s, got %v", errInvalidIP, got)
	}

	if got := ip.Reverse(); got != errInvalidIP.Error() {
		t.Errorf("Reverse() for ipZero; expected %s, got %v", errInvalidIP, got)
	}
}

func TestParseIP(t *testing.T) {
	if ip := mustIP("127.0.0.1"); !ip.Is4() {
		t.Error("127.0.0.1 should be v4")
	}
	if ip := mustIP("fffe::1"); !ip.Is6() {
		t.Error("fffe::1 should be v6")
	}
	if ip := mustIP(net.IP([]byte{1, 2, 3, 4})); !ip.Is4() {
		t.Error("1.2.3.4 should be v4")
	}
	if ip, _ := FromStdIP(net.IP([]byte{1, 2, 3, 4, 5})); ip.IsValid() {
		t.Error("1.2.3.4.5 should be invalid")
	}
}

func TestPanic_1(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	mustIP(1)
}

func TestPanic_2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	// should panic
	var ip IP
	_ = ip.ToStdIP()
}

func TestLess(t *testing.T) {
	ip1 := mustIP("127.0.0.1")
	if ip1.Less(ip1) {
		t.Error("ip.Less(ip) should return false")
	}
	ip2 := mustIP("::1")
	if !ip1.Less(ip2) {
		t.Error("v4 should be less v6")
	}
	if ip2.Less(ip1) {
		t.Error("v6 should be greater as v4")
	}
}

func TestIP_addOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{ipZero, ipZero, false},
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
		if got := tt.in.addOne(); got != tt.want {
			t.Errorf("(%v).addOne = %v; want: %v", tt.in, got, tt.want)
		}
	}

}

func TestIP_subOne(t *testing.T) {
	ips := []struct {
		in   IP
		want IP
		ok   bool
	}{
		{ipZero, ipZero, false},
		{mustIP("0.0.0.0"), ipZero, false},
		{mustIP("0.0.0.1"), mustIP("0.0.0.0"), true},
		{mustIP("127.0.0.1"), mustIP("127.0.0.0"), true},
		//
		{mustIP("::"), ipZero, false},
		{mustIP("::1"), mustIP("::"), true},
		{mustIP("0:0:0:1::"), mustIP("::ffff:ffff:ffff:ffff"), true},
	}

	for _, tt := range ips {
		if got := tt.in.subOne(); got != tt.want {
			t.Errorf("(%v).subOne = %v; want: %v", tt.in, got, tt.want)
		}
	}
}

func TestIP_fromBytes(t *testing.T) {
	tests := []struct {
		in   []byte
		want IP
	}{
		{[]byte{0, 0, 0, 0}, mustIP("0.0.0.0")},
		{[]byte{15: 0}, mustIP("::0")},
		{[]byte{7: 0}, ipZero},
	}

	for _, tt := range tests {
		if got, _ := fromBytes(tt.in); got != tt.want {
			t.Errorf("fromBytes(%v) = %v, want: %v", tt.in, got, tt.want)
		}
	}
}
