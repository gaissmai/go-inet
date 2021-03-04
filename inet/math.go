package inet

import (
	"encoding/binary"
	"math/bits"
)

// For bit calculations convert the IP address to 32 Bits or 2x64 Bits.
type ipAsBits struct {
	version  uint8  // IP version
	v4Bits   uint32 // IPv4
	v6loBits uint64 // IPv6 lower part
	v6hiBits uint64 // IPv6 higher part
}

// IP address octets are BigEndian encoded, uint32 or uint128 (2x uint64, sic)
// cut off the first octet as version tag
func (ip IP) toBits() ipAsBits {
	out := ipAsBits{version: ip.octets[0]}

	if out.version == IPv4 {
		out.v4Bits = binary.BigEndian.Uint32([]byte(ip.octets[1:]))
		return out
	}

	out.v6hiBits = binary.BigEndian.Uint64([]byte(ip.octets[1:9]))
	out.v6loBits = binary.BigEndian.Uint64([]byte(ip.octets[9:17]))
	return out
}

// addOne increments the IP by one, returns (IPZero, false) on overflow
func (ip IP) addOne() (IP, bool) {
	x := ip.toBits()

	if x.version == IPv4 {
		return addOne32(x)
	}

	if x.version == IPv6 {
		return addOne128(x)
	}
	panic(errInvalidIP)
}

func addOne32(x ipAsBits) (IP, bool) {
	out := make([]byte, 4+1)
	var carry uint32
	a := x.v4Bits
	a, carry = bits.Add32(a, 1, 0)

	if carry == 1 {
		return ipZero, false
	}

	out[0] = IPv4
	binary.BigEndian.PutUint32(out[1:], a)

	return IP{string(out)}, true
}

func addOne128(x ipAsBits) (IP, bool) {
	out := make([]byte, 16+1)
	var carry uint64
	lo := x.v6loBits
	hi := x.v6hiBits

	lo, carry = bits.Add64(lo, 1, 0)
	hi, carry = bits.Add64(hi, 0, carry)

	if carry == 1 {
		return ipZero, false
	}

	out[0] = IPv6
	binary.BigEndian.PutUint64(out[1:9], hi)
	binary.BigEndian.PutUint64(out[9:17], lo)

	return IP{string(out)}, true
}

// subOne decrements the IP by one, returns (IPZero, false) on underflow
func (ip IP) subOne() (IP, bool) {
	x := ip.toBits()

	if x.version == IPv4 {
		return subOne32(x)
	}

	if x.version == IPv6 {
		return subOne128(x)
	}
	panic(errInvalidIP)
}

func subOne32(x ipAsBits) (IP, bool) {
	out := make([]byte, 4+1)

	var borrow uint32
	a := x.v4Bits
	a, borrow = bits.Sub32(a, 1, 0)

	if borrow == 1 {
		return ipZero, false
	}

	out[0] = IPv4
	binary.BigEndian.PutUint32(out[1:], a)

	return IP{string(out)}, true
}

func subOne128(x ipAsBits) (IP, bool) {
	out := make([]byte, 16+1)
	var borrow uint64
	lo := x.v6loBits
	hi := x.v6hiBits

	lo, borrow = bits.Sub64(lo, 1, 0)
	hi, borrow = bits.Sub64(hi, 0, borrow)

	if borrow == 1 {
		return ipZero, false
	}

	out[0] = IPv6
	binary.BigEndian.PutUint64(out[1:9], hi)
	binary.BigEndian.PutUint64(out[9:17], lo)

	return IP{string(out)}, true
}

// bitLen returns the common bits as maskLen and the trailing bits as hostLen.
func (a Block) bitLen() (maskLen, hostLen int) {
	base := a.base.toBits()
	last := a.last.toBits()

	// v4
	if base.version == IPv4 {
		// common bits = leadingZeros(a XOR b)
		maskLen = bits.LeadingZeros32(base.v4Bits ^ last.v4Bits)
		hostLen = 32 - maskLen
		return
	}

	// v6
	maskLen = bits.LeadingZeros64(base.v6hiBits ^ last.v6hiBits)
	if maskLen == 64 {
		maskLen += bits.LeadingZeros64(base.v6loBits ^ last.v6loBits)
	}
	hostLen = 128 - maskLen
	return
}
