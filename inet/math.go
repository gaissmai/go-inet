package inet

import (
	"encoding/binary"
	"errors"
	"math/bits"
)

var (
	errOverflow  = errors.New("overflow")
	errUnderflow = errors.New("underflow")
)

// for calculations represent the IP address to a number
// use one struct for both versions
type asUint struct {
	version  int
	bitsv4   uint32 // IPv4
	bitsv6lo uint64 // IPv6 lower part
	bitsv6hi uint64 // IPv6 higher part
}

func (ip IP) toUint() asUint {
	if ip[0] == IPv4 {
		return asUint{version: IPv4, bitsv4: binary.BigEndian.Uint32([]byte(ip[1:]))}
	}
	if ip[0] == IPv6 {
		return asUint{
			version:  IPv6,
			bitsv6hi: binary.BigEndian.Uint64([]byte(ip[1:9])),
			bitsv6lo: binary.BigEndian.Uint64([]byte(ip[9:17])),
		}
	}
	panic(errInvalidIP)
}

// addOne increments the IP by one, returns (IPZero, false) on overflow
func (ip IP) addOne() (ip2 IP, ok bool) {
	out := make([]byte, len(ip))
	x := ip.toUint()

	if x.version == IPv4 {

		var carry uint32
		a := x.bitsv4
		a, carry = bits.Add32(a, 1, 0)

		if carry == 1 {
			return IPZero, false
		}

		out[0] = IPv4
		binary.BigEndian.PutUint32(out[1:], a)

		return IP(string(out)), true
	}

	if x.version == IPv6 {

		var carry uint64
		lo := x.bitsv6lo
		hi := x.bitsv6hi

		lo, carry = bits.Add64(lo, 1, 0)
		hi, carry = bits.Add64(hi, 0, carry)

		if carry == 1 {
			return IPZero, false
		}

		out[0] = IPv6
		binary.BigEndian.PutUint64(out[1:9], hi)
		binary.BigEndian.PutUint64(out[9:17], lo)

		return IP(string(out)), true
	}

	panic(errInvalidIP)
}

// subOne decrements the IP by one, returns (IPZero, false) on underflow
func (ip IP) subOne() (ip2 IP, ok bool) {
	out := make([]byte, len(ip))
	x := ip.toUint()

	if x.version == IPv4 {
		var borrow uint32
		a := x.bitsv4
		a, borrow = bits.Sub32(a, 1, 0)

		if borrow == 1 {
			return IPZero, false
		}

		out[0] = IPv4
		binary.BigEndian.PutUint32(out[1:], a)

		return IP(string(out)), true
	}

	if x.version == IPv6 {
		var borrow uint64
		lo := x.bitsv6lo
		hi := x.bitsv6hi

		lo, borrow = bits.Sub64(lo, 1, 0)
		hi, borrow = bits.Sub64(hi, 0, borrow)

		if borrow == 1 {
			return IPZero, false
		}

		out[0] = IPv6
		binary.BigEndian.PutUint64(out[1:9], hi)
		binary.BigEndian.PutUint64(out[9:17], lo)

		return IP(string(out)), true
	}

	panic(errInvalidIP)
}

// bitLen returns the minimum number of bits to represent the block.
func (a Block) bitLen() int {
	b := a.base.toUint()
	l := a.last.toUint()

	// v4
	if b.version == IPv4 {
		return 32 - bits.LeadingZeros32(b.bitsv4^l.bitsv4)
	}

	// v6
	n := bits.LeadingZeros64(b.bitsv6hi ^ l.bitsv6hi)
	if n == 64 {
		n += bits.LeadingZeros64(b.bitsv6lo ^ l.bitsv6lo)
	}
	return 128 - n
}
