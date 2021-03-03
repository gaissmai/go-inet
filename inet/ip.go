package inet

import (
	"encoding/hex"
	"errors"
	"net"
	"sort"
	"strconv"
)

const (
	// IPv4 version const.
	IPv4 = 4
	// IPv6 version const.
	IPv6 = 6
	// IPZero string constant.
	IPZero = ""
)

var errInvalidIP = errors.New("invalid IP")

// IP represents a single IPv4 or IPv6 address as an opaque string.
// The first byte in the string determines the version.
//
// This IP representation is comparable and can be used as keys in maps
// and fast sorted without conversions to/from the different IP versions.
type IP string

// ParseIP parses and returns the input as type IP.
// The input type may be:
//   string
//   net.IP
//   []byte
//
// The hard part is done by net.ParseIP().
// Returns ipZero and error on invalid input.
func ParseIP(i interface{}) (IP, error) {
	switch v := i.(type) {
	case string:
		return ipFromString(v)
	case net.IP:
		return ipFromNetIP(v)
	case []byte:
		return ipFromBytes(v)
	default:
		return IPZero, errInvalidIP
	}
}

// MustIP is a helper that calls ParseIP and returns just inet.IP or panics on errr.
// It is intended for use in variable initializations.
func MustIP(i interface{}) IP {
	ip, err := ParseIP(i)
	if err != nil {
		panic(err)
	}
	return ip
}

// ipFromString parses s as an IP address, returning the result. The string s can be
// in dotted decimal ("192.0.2.1") or IPv6 ("2001:db8::42") form. If s is not a
// valid textual representation of an IP address, ipFromString returns IP{} and error.
// The real work is done by net.ParseIP() and converted to type IP.
func ipFromString(s string) (IP, error) {
	return ipFromNetIP(net.ParseIP(s))
}

// ipFromNetIP converts from stdlib net.IP ([]byte) to IP opaque string representation.
func ipFromNetIP(netIP net.IP) (IP, error) {
	if netIP == nil {
		return IPZero, errInvalidIP
	}

	if v4 := netIP.To4(); v4 != nil {
		ip := setBytes(v4)
		return ip, nil
	}
	if v6 := netIP.To16(); v6 != nil {
		ip := setBytes(v6)
		return ip, nil
	}
	return IPZero, errInvalidIP
}

// ipFromBytes sets the IP from 4 or 16 bytes. Returns error on wrong number of bytes.
func ipFromBytes(bs []byte) (IP, error) {
	if l := len(bs); l != 4 && l != 16 {
		return IPZero, errInvalidIP
	}
	return setBytes(bs), nil
}

// set the string from []byte input.
func setBytes(bs []byte) IP {
	l := len(bs)
	if l == 4 {
		ip := append([]byte{IPv4}, bs...)
		return IP(string(ip))
	}
	if l == 16 {
		ip := append([]byte{IPv6}, bs...)
		return IP(string(ip))
	}
	panic(errInvalidIP)
}

// bytes returns the ip address in byte representation.
func (ip IP) bytes() []byte {
	return []byte(ip[1:])
}

// ToNetIP converts to net.IP. Panics on invalid input.
func (ip IP) ToNetIP() net.IP {
	return net.IP(ip.bytes())
}

// Version returns 4 or 6. For IPzero it returns 0.
func (ip IP) Version() int {
	if ip == IPZero {
		return 0
	}
	return int(ip[0])
}

// SortIP sorts the given slice in place.
// IPv4 addresses are sorted 'naturally' before IPv6 addresses, no prior conversion or version split necessary.
func SortIP(ips []IP) {
	sort.Slice(ips, func(i, j int) bool { return ips[i] < ips[j] })
}

// Expand IP address into canonical form, useful for grep, aligned output and lexical sort.
func (ip IP) Expand() string {
	if ip[0] == IPv4 {
		return expandIPv4(ip.bytes())
	}

	if ip[0] == IPv6 {
		return expandIPv6(ip.bytes())
	}

	panic(errInvalidIP)
}

//  127.0.0.1 -> 127.000.000.001
func expandIPv4(ip []byte) string {
	out := make([]byte, 0, len(ip)*3+3)

	for i := 0; i < len(ip); i++ {
		if i > 0 {
			out = append(out, '.')
		}

		// Itoa and fixed fast leftpad; much faster than fmt.Sprintf("%03d", i)
		// beware, maybe we handle zig millions of IP addresses.
		s := strconv.Itoa(int(ip[i]))

		// preset and left padding
		leftpad := []byte("000")
		copy(leftpad[3-len(s):], s)

		out = append(out, leftpad...)
	}
	return string(out)
}

// 2001:db8::1 -> 2001:0db8:0000:0000:0000:0000:0000:0001
func expandIPv6(ip []byte) string {
	buf := make([]byte, hex.EncodedLen(len(ip)))
	hex.Encode(buf, ip)

	out := make([]byte, 0, len(buf)+7) // insert 7 x ':'

	for i := 0; i < len(buf); i++ {
		// insert ':' after any hex nibble
		if i > 0 && i%4 == 0 {
			out = append(out, ':')
		}

		out = append(out, buf[i])
	}
	return string(out)
}

// Reverse IP address, needed for PTR entries in DNS zone files.
func (ip IP) Reverse() string {
	if ip[0] == IPv4 {
		return reverseIPv4(ip.bytes())
	}

	if ip[0] == IPv6 {
		return reverseIPv6(ip.bytes())
	}

	panic(errInvalidIP)
}

// []byte{127,0,0,1}} -> "1.0.0.127"
func reverseIPv4(ip []byte) string {
	out := make([]byte, 0, len(ip)*3+3)

	// reverse loop
	for i := len(ip) - 1; i >= 0; i-- {
		s := strconv.Itoa(int(ip[i]))
		out = append(out, []byte(s)...)

		// add sep, but not at loop end
		if i > 0 {
			out = append(out, '.')
		}
	}
	return string(out)
}

// []byte{0xfe, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xa, 0x0, 0x1}
//  -> "1.0.0.0.a.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.e.f"
func reverseIPv6(ip []byte) string {
	buf := make([]byte, hex.EncodedLen(len(ip)))
	hex.Encode(buf, ip)

	out := make([]byte, 0, len(buf)+31) // insert 31 x '.'

	// reverse loop
	for i := len(buf) - 1; i >= 0; i-- {
		out = append(out, buf[i])

		// add sep, but not at loop end
		if i > 0 {
			out = append(out, '.')
		}
	}

	return string(out)
}
