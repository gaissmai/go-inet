package inet

import (
	"encoding/hex"
	"errors"
	"net"
	"strconv"
)

const (
	v4 = 4
	v6 = 6
)

// IP represents a single IPv4 or IPv6 address.
type IP struct {
	version uint8
	uint128 // hi, lo uint64
}

const invalidIP = "invalid IP"

var errInvalidIP = errors.New(invalidIP)

// ParseIP parses and returns the input as type IP.
// Returns the zero value for IP and error on invalid input.
//
// The string form can be in IPv4 dotted decimal ("192.168.2.1"), IPv6
// ("2001:db8::affe"), or IPv4-mapped IPv6 ("::ffff:172.16.0.1").
//
// The hard part is done by net.ParseIP().
func ParseIP(s string) (ip IP, err error) {
	return FromStdIP(net.ParseIP(s))
}

// FromStdIP returns an IP from the standard library's IP type.
//
// If std is <nil>, returns the zero value and error.
func FromStdIP(std net.IP) (ip IP, err error) {
	if std == nil {
		err = errInvalidIP
		return
	}

	if bs := std.To4(); bs != nil {
		return fromBytes(bs)
	}
	if bs := std.To16(); bs != nil {
		return fromBytes(bs)
	}
	err = errInvalidIP
	return
}

// toStdIP converts to net.IP. Panics on invalid input.
func (ip IP) toStdIP() net.IP {
	if !ip.IsValid() {
		panic("toStdIP() called on zero value")
	}
	return net.IP(ip.toBytes())
}

// IsValid reports whether ip is a valid address and not the zero value of the IP type.
// The zero value is not a valid IP address of any type.
//
// Note that "0.0.0.0" and "::" are not the zero value.
func (ip IP) IsValid() bool {
	return ip != IP{}
}

// Is4 reports whether ip is an IPv4 address.
//
// There is no Is4in6. IPv4-mapped IPv6 addresses are stripped down to IPv4 otherwise the sort order would be undefined.
func (ip IP) Is4() bool {
	return ip.version == v4
}

// Is6 reports whether ip is an IPv6 address.
//
// There is no Is4in6. IPv4-mapped IPv6 addresses are stripped down to IPv4 otherwise the sort order would be undefined.
func (ip IP) Is6() bool {
	return ip.version == v6
}

// String returns the string form of the IP address.
// It returns one of 3 forms:
//
//   "invalid IP"  if ip.IsValid() is false
//   "127.0.0.1"
//   "2001:db8::1"
func (ip IP) String() string {
	if !ip.IsValid() {
		return invalidIP
	}
	return ip.toStdIP().String()
}

// Less reports whether the ip should sort before ip2.
// IPv4 addresses sorts always before IPv6 addresses.
func (ip IP) Less(ip2 IP) bool {
	if ip.version != ip2.version {
		return ip.version < ip2.version
	}
	if ip.hi != ip2.hi {
		return ip.hi < ip2.hi
	}
	return ip.lo < ip2.lo
}

// Expand IP address into canonical form, useful for grep, aligned output and lexical sort.
func (ip IP) Expand() string {
	if ip.version == v4 {
		return expandIPv4(ip.toBytes())
	}
	if ip.version == v6 {
		return expandIPv6(ip.toBytes())
	}
	return errInvalidIP.Error()
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
	if ip.version == v4 {
		return reverseIPv4(ip.toBytes())
	}
	if ip.version == v6 {
		return reverseIPv6(ip.toBytes())
	}
	return errInvalidIP.Error()
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
