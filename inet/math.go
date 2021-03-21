package inet

import (
	"encoding/binary"
	"math/bits"
)

// simulate uint128(ipv6) and uint32(ipv4)
type uint128 struct {
	hi uint64
	lo uint64
}

var maxV4 = uint128{uint64(0), uint64(^uint32(0))} //  32x1 :0x00_00..00_ff_ff_ff_ff
var maxV6 = uint128{^uint64(0), ^uint64(0)}        // 128x1 :0xff_ff..ff_ff_ff_ff_ff

// bitwise NOT: ^u
func not(u uint128) uint128 {
	return uint128{^u.hi, ^u.lo}
}

// bitwise OR: u|m
func (u uint128) or(m uint128) uint128 {
	return uint128{u.hi | m.hi, u.lo | m.lo}
}

// bitwise XOR: u^m
func (u uint128) xor(m uint128) uint128 {
	return uint128{u.hi ^ m.hi, u.lo ^ m.lo}
}

// bitwise AND: u&m
func (u uint128) and(m uint128) uint128 {
	return uint128{u.hi & m.hi, u.lo & m.lo}
}

// cmp
func (u uint128) cmp(m uint128) int {
	if u == m {
		return 0
	}
	if u.hi == m.hi {
		if u.lo < m.lo {
			return -1
		}
		return 1
	}
	if u.hi < m.hi {
		return -1
	}
	return 1
}

// strip high 96 bits
func (ip IP) strip96() IP {
	ip.uint128 = uint128{0, uint64(uint32(ip.lo))}
	return ip
}

// fromBytes makes the IP{} from network ordered byte slice
func fromBytes(bs []byte) (IP, error) {
	l := len(bs)
	if l == 4 {
		ip := IP{
			v4,
			uint128{0, uint64(binary.BigEndian.Uint32(bs[:]))},
		}
		return ip, nil
	}

	if l == 16 {
		ip := IP{
			v6,
			uint128{binary.BigEndian.Uint64(bs[:8]), binary.BigEndian.Uint64(bs[8:])},
		}
		return ip, nil
	}

	return IP{}, errInvalidIP
}

// toBytes returns the ip address in network ordered byte representation.
func (ip IP) toBytes() []byte {
	if ip.version == v4 {
		bs := make([]byte, 4)
		binary.BigEndian.PutUint32(bs[:], uint32(ip.lo))
		return bs
	}

	bs := make([]byte, 16)
	binary.BigEndian.PutUint64(bs[:8], ip.hi)
	binary.BigEndian.PutUint64(bs[8:], ip.lo)
	return bs
}

// mkBaseIP makes base address from address and netmask.
//
// base = address & netMask
func (ip IP) mkBaseIP(mask uint128) IP {
	ip.uint128 = ip.and(mask)
	return ip
}

// mkLastIP makes last IP address from base IP address and netmask.
//
// last = base | hostmask
func (ip IP) mkLastIP(mask uint128) IP {
	ip.uint128 = ip.or(not(mask))
	if ip.version == v4 {
		ip = ip.strip96()
	}
	return ip
}

// commonPrefixLen128
func (u uint128) commonPrefixLen128(v uint128) (n uint8) {
	u = u.xor(v)
	if n = uint8(bits.LeadingZeros64(u.hi)); n == 64 {
		n += uint8(bits.LeadingZeros64(u.lo))
	}
	return
}

// commonPrefixLen returns the common prefix len of two IP addresses.
func (ip IP) commonPrefixLen(ip2 IP) uint8 {
	return ip.commonPrefixLen128(ip2.uint128)
}

// isCIDR returns true if the IP pair forms a CIDR.
func (ip IP) isCIDR(ip2 IP) bool {

	// count common prefix bits ... and
	// get mask from pre calculated lookup table
	mask := maskUint128[ip.commonPrefixLen(ip2)]

	// check if mask applied to base and last results
	// in all zeros and all ones
	allZeroBase := ip.xor(ip.and(mask)) == uint128{}
	allOnesLast := ip2.or(mask) == uint128{^uint64(0), ^uint64(0)}

	return allZeroBase && allOnesLast
}

// addOne32 increments the ipv4 address by one, returns ok=false on overflow
func addOne32(u uint128) (m uint128, ok bool) {
	lo, carry := bits.Add32(uint32(u.lo), 1, 0)
	m.lo = uint64(lo)
	if carry == 0 {
		ok = true
	}
	return
}

// addOne128 increments the ipv6 address by one, returns ok=false on overflow
func addOne128(u uint128) (m uint128, ok bool) {
	var carry uint64
	m.lo, carry = bits.Add64(u.lo, 1, 0)
	m.hi, carry = bits.Add64(u.hi, 0, carry)
	if carry == 0 {
		ok = true
	}
	return
}

// subOne32 decrements the ipv4 address by one, returns ok=false on underflow
func subOne32(u uint128) (m uint128, ok bool) {
	lo, borrow := bits.Sub32(uint32(u.lo), 1, 0)
	m.lo = uint64(lo)
	if borrow == 0 {
		ok = true
	}
	return
}

// subOne128 decrements the ipv6 address by one, returns ok=false on underflow
func subOne128(u uint128) (m uint128, ok bool) {
	var borrow uint64
	m.lo, borrow = bits.Sub64(u.lo, 1, 0)
	m.hi, borrow = bits.Sub64(u.hi, 0, borrow)
	if borrow == 0 {
		ok = true
	}
	return
}

// addOne increments the IP by one, returns IPZero on overflow.
func (ip IP) addOne() IP {
	if !ip.IsValid() {
		return IP{}
	}

	var ok bool
	if ip.version == v4 {
		ip.uint128, ok = addOne32(ip.uint128)
	} else {
		ip.uint128, ok = addOne128(ip.uint128)
	}
	if !ok {
		return IP{}
	}
	return ip
}

// subOne decrements the IP by one, returns IPZero on underflow.
func (ip IP) subOne() IP {
	if !ip.IsValid() {
		return IP{}
	}

	var ok bool
	if ip.version == v4 {
		ip.uint128, ok = subOne32(ip.uint128)
	} else {
		ip.uint128, ok = subOne128(ip.uint128)
	}
	if !ok {
		return IP{}
	}
	return ip
}
