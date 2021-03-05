package inet

import (
	"encoding/binary"
	"math/bits"
)

// simulate uint128(ipv6) and uint32(ipv4)
type address struct {
	hi uint64
	lo uint64
}

// bitwise NOT: ^u
func not(u address) address {
	return address{^u.hi, ^u.lo}
}

// bitwise OR: u|m
func or(u address, m address) address {
	return address{u.hi | m.hi, u.lo | m.lo}
}

// bitwise XOR: u^m
func xor(u address, m address) address {
	return address{u.hi ^ m.hi, u.lo ^ m.lo}
}

// bitwise AND: u&m
func and(u address, m address) address {
	return address{u.hi & m.hi, u.lo & m.lo}
}

// cmp
func cmp(u address, m address) int {
	if u == m {
		return 0
	}
	if u.hi == m.hi {
		if u.lo < m.lo {
			return -1
		} else {
			return 1
		}
	}
	if u.hi < m.hi {
		return -1
	} else {
		return 1
	}
}

// bitwise operators work on 128 bits,
// IP address operators must divide between ipv4 and ipv6
func (ip IP) stripTov4() IP {
	ip.hi = 0
	ip.lo = uint64(uint32(ip.lo))
	return ip
}

// not is the bitwise inverse of address
func (ip IP) not() IP {
	ip.address = not(ip.address)
	if ip.version == ipv4 {
		ip = ip.stripTov4()
	}
	return ip
}

// or is the bitwise or: ip.address | mask.address
func (ip IP) or(mask IP) IP {
	ip.address = or(ip.address, mask.address)
	if ip.version == ipv4 {
		ip = ip.stripTov4()
	}
	return ip
}

// xor is the bitwise xor: ip.address ^ mask.address
func (ip IP) xor(mask IP) IP {
	ip.address = xor(ip.address, mask.address)
	if ip.version == ipv4 {
		ip = ip.stripTov4()
	}
	return ip
}

// and is the bitwise and: ip.address & mask.address
func (ip IP) and(mask IP) IP {
	ip.address = and(ip.address, mask.address)
	if ip.version == ipv4 {
		ip = ip.stripTov4()
	}
	return ip
}

// cmp
func (ip IP) cmp(ip2 IP) int {
	if ip.version < ip2.version {
		return -1
	}
	if ip.version > ip2.version {
		return 1
	}
	return cmp(ip.address, ip2.address)
}

// addOne32 increments the ipv4 address by one, returns ok=false on overflow
func addOne32(u address) (m address, ok bool) {
	lo, carry := bits.Add32(uint32(u.lo), 1, 0)
	m.lo = uint64(lo)
	if carry == 0 {
		ok = true
	}
	return
}

// addOne128 increments the ipv6 address by one, returns ok=false on overflow
func addOne128(u address) (m address, ok bool) {
	var carry uint64
	m.lo, carry = bits.Add64(u.lo, 1, 0)
	m.hi, carry = bits.Add64(u.hi, 0, carry)
	if carry == 0 {
		ok = true
	}
	return
}

// subOne32 decrements the ipv4 address by one, returns ok=false on overflow
func subOne32(u address) (m address, ok bool) {
	lo, borrow := bits.Sub32(uint32(u.lo), 1, 0)
	m.lo = uint64(lo)
	if borrow == 0 {
		ok = true
	}
	return
}

// subOne128 decrements the ipv6 address by one, returns ok=false on overflow
func subOne128(u address) (m address, ok bool) {
	var borrow uint64
	m.lo, borrow = bits.Sub64(u.lo, 1, 0)
	m.hi, borrow = bits.Sub64(u.hi, 0, borrow)
	if borrow == 0 {
		ok = true
	}
	return
}

// addOne increments the IP by one, returns (IPZero, false) on overflow
func (ip IP) addOne() (ip2 IP, ok bool) {
	if ip.IsZero() {
		panic("addOne() called on zero value")
	}
	ip2 = ip

	if ip.version == ipv4 {
		ip2.address, ok = addOne32(ip.address)
	} else {
		ip2.address, ok = addOne128(ip.address)
	}
	if !ok {
		return ipZero, false
	}
	return ip2, true
}

// subOne decrements the IP by one, returns (IPZero, false) on underflow
func (ip IP) subOne() (ip2 IP, ok bool) {
	if ip.IsZero() {
		panic("addOne() called on zero value")
	}
	ip2 = ip

	if ip.version == ipv4 {
		ip2.address, ok = subOne32(ip.address)
	} else {
		ip2.address, ok = subOne128(ip.address)
	}
	if !ok {
		return ipZero, false
	}
	return ip2, true
}

// fromBytes makes the IP{} from network ordered byte slice
func fromBytes(bs []byte) (IP, error) {
	l := len(bs)
	if l == 4 {
		ip := IP{
			ipv4,
			address{0, uint64(binary.BigEndian.Uint32(bs[:]))},
		}
		return ip, nil
	}

	if l == 16 {
		ip := IP{
			ipv6,
			address{binary.BigEndian.Uint64(bs[:8]), binary.BigEndian.Uint64(bs[8:])},
		}
		return ip, nil
	}

	return ipZero, errInvalidIP
}

// toBytes returns the ip address in network ordered byte representation.
func (ip IP) toBytes() []byte {
	if ip.version == ipv4 {
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
func mkBaseIP(any IP, mask IP) IP {
	out := any.and(mask)
	return out
}

// mkLastIP makes last IP address from base IP address and netmask.
//
// last = base | hostmask
func mkLastIP(base IP, mask IP) IP {
	out := base.or(mask.not())
	return out
}

// prefixLen64
func prefixLen64(u, v uint64) uint8 {
	return uint8(bits.LeadingZeros64(u ^ v))
}

// prefixLen128
func prefixLen128(u, v address) (n uint8) {
	if n = prefixLen64(u.hi, v.hi); n == 64 {
		n += prefixLen64(u.lo, v.lo)
	}
	return
}

// commonPrefixLen returns the common prefix len of base and last IP in block.
func (b Block) commonPrefixLen() uint8 {
	return prefixLen128(b.base.address, b.last.address)
}

// commonPrefixMask returns the common prefix bitmask of base and last IP in block.
func (b Block) commonPrefixMask() (mask IP, ok bool) {
	mask.version = b.base.version

	n := b.commonPrefixLen()
	if n > 64 {
		mask.address = address{^uint64(0), ^uint64(0) << (128 - n)}
	} else {
		mask.address = address{^uint64(0) << (64 - n), 0}
	}
	if mask.version == ipv4 {
		mask = mask.stripTov4()
	}

	// base & netmask = base AND base | hostmask = last
	if b.base.and(mask) == b.base && b.base.or(mask.not()) == b.last {
		return mask, true
	}
	return ipZero, false
}
