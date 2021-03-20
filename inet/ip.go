package inet

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"net"
	"sort"
	"strconv"
	"strings"
)

var (
	errInvalidIP = errors.New("invalid IP")
	errOverflow  = errors.New("overflow")
	errUnderflow = errors.New("underflow")
)

// IP represents a single IPv4 or IPv6 address as string.
// The first byte in the string has the value 0x04 or 0x06
//
// This IP representation is comparable and can be used as key in maps
// and fast sorted without conversions to/from the different IP versions.
type IP string

// ParseIP parses and returns the input as type IP.
// The input type may be:
//   string
//   net.IP
//   []byte
//
// The hard part is done by net.ParseIP().
// Returns IP{} and error on invalid input.
func ParseIP(i interface{}) (IP, error) {
	switch v := i.(type) {
	case string:
		return ipFromString(v)
	case net.IP:
		return ipFromNetIP(v)
	case []byte:
		return ipFromBytes(v)
	default:
		return "", errInvalidIP
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
		return "", errInvalidIP
	}

	if v4 := netIP.To4(); v4 != nil {
		ip := setBytes(v4)
		return ip, nil
	}
	if v6 := netIP.To16(); v6 != nil {
		ip := setBytes(v6)
		return ip, nil
	}
	return "", errInvalidIP
}

// ipFromBytes sets the IP from 4 or 16 bytes. Returns error on wrong number of bytes.
func ipFromBytes(bs []byte) (IP, error) {
	if l := len(bs); l != 4 && l != 16 {
		return "", errInvalidIP
	}
	return setBytes(bs), nil
}

// set the string from []byte input.
func setBytes(bs []byte) IP {
	l := len(bs)
	if l == 4 {
		return IP("\x04" + string(bs))
	}
	if l == 16 {
		return IP("\x06" + string(bs))
	}
	panic(errInvalidIP)
}

// Bytes returns the ip address in byte representation. Returns 4 bytes for IPv4 and 16 bytes for IPv6.
// Panics on invalid input.
func (ip IP) Bytes() []byte {
	if ip[0] == 4 || ip[0] == 6 {
		return []byte(ip[1:])
	}
	panic(errInvalidIP)
}

// ToNetIP converts to net.IP. Panics on invalid input.
func (ip IP) ToNetIP() net.IP {
	return net.IP(ip.Bytes())
}

// IsValid returns true on valid IPs, false otherwise.
func (ip IP) IsValid() bool {
	// wrong length?
	l := len(ip)
	if l != 5 && l != 17 {
		return false
	}
	// wrong version?
	if ip[0] != 4 && ip[0] != 6 {
		return false
	}
	return true
}

// Version returns 4 or 6 for valid IPs. Panics on invalid IP.
func (ip IP) Version() int {
	if ip[0] != 4 && ip[0] != 6 {
		panic(errInvalidIP)
	}

	return int(ip[0])
}

// Compare returns an integer comparing two IP addresses lexicographically. The
// result will be:
//   0 if a == b
//  -1 if a < b
//  +1 if a > b
//
// Also the string comparison operators are possible.
// IPv4 addresses are always less than IPv6 addresses.
func (ip IP) Compare(ip2 IP) int {
	return strings.Compare(string(ip), string(ip2))
}

// SortIP sorts the given slice in place.
// IPv4 addresses are sorted 'naturally' before IPv6 addresses, no prior conversion or version split necessary.
func SortIP(ips []IP) {
	sort.Slice(ips, func(i, j int) bool { return ips[i] < ips[j] })
}

// Expand IP address into canonical form, useful for grep, aligned output and lexical sort.
func (ip IP) Expand() string {
	if ip[0] == 4 {
		return expandIPv4(ip.Bytes())
	}

	if ip[0] == 6 {
		return expandIPv6(ip.Bytes())
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
	if ip[0] == 4 {
		return reverseIPv4(ip.Bytes())
	}

	if ip[0] == 6 {
		return reverseIPv6(ip.Bytes())
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

// AddUint64 adds i to ip, panics on overflow.
func (ip IP) AddUint64(i uint64) IP {

	// convert i to bytes, forward to AddBytes
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf[:], i)

	return ip.AddBytes(buf)
}

// AddBytes adds byte slice to ip, panics on overflow.
func (ip IP) AddBytes(bs []byte) IP {
	// get the IP address as []byte slice
	ipAsBytes := ip.Bytes()

	y := new(big.Int).SetBytes(bs)
	z := new(big.Int).SetBytes(ipAsBytes)

	z.Add(z, y)

	// get the big.Int as []byte slice
	zbs := z.Bytes()

	// overflow?
	if len(zbs) > len(ipAsBytes) {
		panic(errOverflow)
	}

	// left padding with zeros
	// 1 => []byte{0,0,0,1} for IPv4
	// 1 => []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1} for IPv6
	leftpad := make([]byte, len(ipAsBytes))
	copy(leftpad[len(leftpad)-len(zbs):], zbs)

	return setBytes(leftpad)
}

// SubUint64 subtracts i from ip, panics on underflow.
func (ip IP) SubUint64(i uint64) IP {

	// convert to bytes, forward to SubBytes
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf[:], i)

	return ip.SubBytes(buf)
}

// SubBytes subtract byte slice from ip, panics on underflow.
func (ip IP) SubBytes(bs []byte) IP {
	// get the IP address as []byte slice
	ipAsBytes := ip.Bytes()

	y := new(big.Int).SetBytes(bs)
	z := new(big.Int).SetBytes(ipAsBytes)

	z.Sub(z, y)

	// underflow?
	bigZero := new(big.Int)
	if z.Cmp(bigZero) == -1 {
		panic(errUnderflow)
	}

	// get the big.Int as []byte slice
	zIPAsBytes := z.Bytes()

	// left padding with zeros
	// 1 => []byte{0,0,0,1} for IPv4
	// 1 => []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1} for IPv6
	leftpad := make([]byte, len(ipAsBytes))
	copy(leftpad[len(leftpad)-len(zIPAsBytes):], zIPAsBytes)

	return setBytes(leftpad)
}
