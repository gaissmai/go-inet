package inet

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
)

var (
	errInvalidBlock = errors.New("invalid Block")
	errOverflow     = errors.New("overflow")
	errUnderflow    = errors.New("underflow")
)

// Block is an IP-network or IP-range, e.g.
//
//  192.168.0.1/24              // network, with CIDR mask
//  ::1/128                     // network, with CIDR mask
//  10.0.0.3-10.0.17.134        // range
//  2001:db8::1-2001:db8::f6    // range
//
// This Block representation is comparable and can be used as key in maps
// and fast sorted without conversions to/from the different IP versions.
type Block struct {
	base IP
	last IP
	mask IP // IP{} for ranges without CIDR mask
}

// the zero-value for type Block, not public
var blockZero Block

// Base returns the blocks base IP address.
func (a Block) Base() IP { return a.base }

// Last returns the blocks last IP address.
func (a Block) Last() IP { return a.last }

// Mask returns the blocks mask as IP address. May be IPZero for begin-end ranges
func (a Block) Mask() IP { return a.mask }

var (
	// MaxCIDRSplit limits the input parameter for SplitCIDR() to 20 (max 2^20 CIDRs) for security.
	MaxCIDRSplit int = 20

	// internal for overflow checks in aggregates
	ipMaxV4 = IP{string([]byte{4, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})}
	ipMaxV6 = IP{string([]byte{6, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})}
)

// ParseBlock parses and returns the input as type Block.
// The input type may be:
//
//   string
//   inet.IP
//   net.IP
//   net.IPNet
//
// Example for valid input strings:
//
//  "2001:db8:dead::/38"
//  "10.0.0.0/8"
//  "4.4.4.4"
//
//  "2001:db8::1-2001:db8::ff00:35"
//  "192.168.2.3-192.168.7.255"
//
// If a begin-end range can be represented as a CIDR, ParseBlock() generates the netmask
// and returns the range as CIDR.
//
// IP addresses as input are converted to /32 or /128 blocks
// Returns error and Block{} on invalid input.
func ParseBlock(i interface{}) (Block, error) {
	switch v := i.(type) {
	case string:
		return blockFromString(v)
	case IP:
		return blockFromIP(v)
	case net.IP:
		return blockFromNetIP(v)
	case net.IPNet:
		return blockFromNetIPNet(v)
	default:
		return blockZero, errInvalidBlock
	}
}

// MustBlock is a helper that calls ParseBlock and returns just inet.Block or panics on error.
// It is intended for use in variable initializations.
func MustBlock(i interface{}) Block {
	b, err := ParseBlock(i)
	if err != nil {
		panic(err)
	}
	return b
}

// blockFromString parses s in network CIDR or in begin-end IP address-range notation.
func blockFromString(s string) (Block, error) {
	if s == "" {
		return blockZero, errInvalidBlock
	}

	i := strings.IndexByte(s, '/')
	if i >= 0 {
		return blockFromCIDR(s)
	}

	i = strings.IndexByte(s, '-')
	if i >= 0 {
		return blockFromRange(s, i)
	}

	// maybe just an ip
	ip, err := ParseIP(s)
	if err == nil {
		return blockFromIP(ip)
	}

	return blockZero, errInvalidBlock
}

// blockFromIP converts inet.IP to inet.Block with /32 or /128 CIDR mask
func blockFromIP(ip IP) (Block, error) {
	b := Block{base: ip, last: ip}
	b.mask = b.getMask()
	return b, nil
}

// blockFromNetIP converts net.IP to inet.Block with /32 or /128 CIDR mask
func blockFromNetIP(nip net.IP) (Block, error) {
	ip, err := FromStdIP(nip)
	if err != nil {
		return blockZero, err
	}
	return blockFromIP(ip)
}

// blockFromNetIPNet converts from stdlib net.IPNet to ip.Block representation.
func blockFromNetIPNet(ipnet net.IPNet) (Block, error) {
	var err error
	a := blockZero

	a.base, err = FromStdIP(ipnet.IP)
	if err != nil {
		return blockZero, errInvalidBlock
	}

	a.mask, err = FromStdIP(net.IP(ipnet.Mask)) // cast needed
	if err != nil {
		return blockZero, errInvalidBlock
	}

	a.last = lastIP(a.base, a.mask)

	return a, nil
}

// parse IP CIDR
// e.g.: 127.0.0.0/8 or 2001:db8::/32
func blockFromCIDR(s string) (Block, error) {
	if strings.HasPrefix(s, "::ffff:") && strings.IndexByte(s, '.') > 6 {
		s = s[7:]
	}
	_, netIPNet, err := net.ParseCIDR(s)
	if err != nil {
		return blockZero, err
	}

	return blockFromNetIPNet(*netIPNet)
}

// lastIP makes last IP address from base IP address and netmask.
//
// last = base | hostMask
//
// Example:
//   ^netMask(255.0.0.0) = hostMask(0.255.255.255)
//
//   ^0xff_00_00_00  = 0x00_ff_ff_ff
//  -----------------------------------------------
//
//    0x7f_00_00_00 base
//  | 0x00_ff_ff_ff hostMask
//  ----------------------
//    0x7f_ff_ff_ff last
//
func lastIP(base IP, mask IP) IP {
	b := base.bytes()
	m := mask.bytes()

	last := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		last[i] = b[i] | ^m[i]
	}
	return setBytes(last)
}

// baseIP makes base address from address and netmask:
//  base[i] = address[i] & netMask[i]
func baseIP(any IP, mask IP) IP {
	a := any.bytes()
	m := mask.bytes()

	base := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		base[i] = a[i] & m[i]
	}
	return setBytes(base)
}

// parse IP address-range
// e.g.: 127.0.0.0-127.0..0.17 or 2001:db8::1-2001:dbb::ffff
func blockFromRange(s string, i int) (Block, error) {
	// split string
	base, last := s[:i], s[i+1:]

	baseIP, err := ParseIP(base)
	if err != nil {
		return blockZero, errInvalidBlock
	}

	lastIP, err := ParseIP(last)
	if err != nil {
		return blockZero, errInvalidBlock
	}

	// begin-end have version mismatch
	if baseIP.octets[0] != lastIP.octets[0] {
		return blockZero, errInvalidBlock
	}

	// begin > end
	if baseIP.octets > lastIP.octets {
		return blockZero, errInvalidBlock
	}

	a := Block{base: baseIP, last: lastIP}

	// maybe range is true CIDR, try to get also the mask
	a.mask = a.getMask()

	return a, nil
}

// String implements the fmt.Stringer interface.
func (a Block) String() string {
	if a.mask == ipZero {
		return fmt.Sprintf("%s-%s", a.base, a.last)
	}

	mb := a.mask.bytes()
	ones, _ := net.IPMask(mb).Size()
	return fmt.Sprintf("%s/%d", a.base, ones)
}

// Covers reports whether Block a contains Block b. a and b may NOT coincide.
// Covers returns true when Block a is a *true* cover of Block b, Equal must then be false.
// If Covers is true then Less must also be true.
//
//  a |-----------------| |-----------------| |-----------------|
//  b   |------------|    |------------|           |------------|
func (a Block) Covers(b Block) bool {
	if a == b {
		return false
	}
	return a.base.octets <= b.base.octets && a.last.octets >= b.last.octets
}

// Less reports whether the Block a should be sorted before Block b.
// REMEMBER: sort containers always to the left.
//
//  a |---|
//  b        |------|
//
//  a |-------|
//  b    |------------|
//
//  a |-----------------|
//  b    |----------|
//
//  a |-----------------|
//  b |------------|
func (a Block) Less(b Block) bool {
	if a.base == b.base {
		// a.Base == b.Base and a covers b, REMEMBER: sort containers to the left
		return a.last.octets > b.last.octets
	}
	return a.base.octets < b.base.octets
}

// IsCIDR returns true if the block is a true CIDR, not just a begin-end range.
func (a Block) IsCIDR() bool {
	return a.mask != ipZero
}

// IsDisjunctWith reports whether the Blocks a and b are disjunct
//  a       |----------|
//  b |---|
//
//  a |------|
//  b          |---|
func (a Block) IsDisjunctWith(b Block) bool {

	//  a       |----------|
	//  b |---|
	if a.base.octets > b.last.octets {
		return true
	}

	//  a |------|
	//  b          |---|
	if a.last.octets < b.base.octets {
		return true
	}

	return false
}

// OverlapsWith reports whether the Blocks overlaps.
//
//  a    |-------|
//  b |------|
//
//  a |------|
//  b    |-------|
//
//  a |----|
//  b      |---------|
//
//  a      |---------|
//  b |----|
func (a Block) OverlapsWith(b Block) bool {
	if a == b {
		return false
	}
	if a.Covers(b) || b.Covers(a) {
		return false
	}
	if a.IsDisjunctWith(b) {
		return false
	}
	return true
}

// SplitCIDR returns the next 2^n CIDRs, splitted from outer block.
// The number of CIDRs is limited to MaxCIDRSplit, panics if more CIDRs are requested.
// Returns nil at max mask length or if block is no CIDR.
func (a Block) SplitCIDR(n int) []Block {
	// algorithm:
	// - create new mask
	// loop
	// - make next base and next last with new mask
	// - break if next last == a.last
	// - ... or increment next last, use it as new base
	// end

	// limit cpu and memory
	if n > MaxCIDRSplit {
		panic("too many CIDRs requested")
	}

	if a == blockZero {
		return nil
	}

	if !a.IsCIDR() {
		return nil
	}

	// check for max mask len, bits are 32 or 128 (v4 or v6)
	maskSize, bits := net.IPMask(a.mask.bytes()).Size()
	if n <= 0 || maskSize+n > bits {
		return nil
	}

	newMask := setBytes(net.CIDRMask(maskSize+n, bits))

	cidrs := make([]Block, 0, 1<<uint(n))

	base := a.base
	if baseIP(base, newMask) != base {
		panic("logic error ...")
	}

	for {
		next := blockZero
		next.base = baseIP(base, newMask)
		next.last = lastIP(next.base, newMask)
		next.mask = newMask
		cidrs = append(cidrs, next)

		// last of outer CIDR already reached?
		if next.last.octets < a.last.octets {
			base, _ = next.last.addOne() // next base
		} else if next.last == a.last {
			break
		} else {
			panic("logic error...")
		}
	}
	return cidrs
}

// FindFreeCIDR returns all free CIDR blocks (of max possible bitlen) within given CIDR,
// minus the inner CIDR blocks.
// Panics if inner blocks are no subset of (or not equal to) outer block.
func (a Block) FindFreeCIDR(bs []Block) []Block {
	for _, i := range bs {
		if !(a.Covers(i) || i == a) {
			panic("at least one inner block isn't contained in (or equal to) outer block")
		}
	}

	free := make([]Block, 0, 10)

	candidates := make([]Block, 0, 10)
	candidates = append(candidates, a) // start with outer block

	for i := 0; i < len(candidates); i++ {
		c := candidates[i]

		if c == blockZero {
			continue
		}

		// hit
		if c.isDisjunctWithAll(bs) {
			free = append(free, c)
			continue
		}

		// c is already a subset of an inner block, don't split further!
		if c.isSubsetOfAny(bs) {
			continue
		}

		// split one bit further, maybe a smaller CIDR is free
		splits := c.SplitCIDR(1)
		candidates = append(candidates, splits...)

		// limit cpu and memory
		if len(candidates) > 1<<uint(MaxCIDRSplit) {
			panic("too many CIDRs generated")
		}
	}

	if len(free) == 0 {
		return nil
	}

	sort.Slice(free, func(i, j int) bool { return free[i].Less(free[j]) })
	return free
}

// isDisjunctWithAll is a helper method. Returns true if a is disjunct with any inner block
func (a Block) isDisjunctWithAll(bs []Block) bool {
	for _, b := range bs {
		if !a.IsDisjunctWith(b) {
			return false
		}
	}
	return true
}

// isSubsetOfAny is a helper method: Returns true if a is subset of any inner block
func (a Block) isSubsetOfAny(bs []Block) bool {
	for _, b := range bs {
		if b.Covers(a) {
			return true
		}
	}
	return false
}

// getMask is a helper method. Calculate a netmask from begin-end, returns ipZero if not possible.
func (a Block) getMask() (mask IP) {
	// v4 or v6
	bits := 32
	if a.base.octets[0] == IPv6 {
		bits = 128
	}

	netLen, _ := a.bitLen()

	mask = setBytes(net.CIDRMask(netLen, bits))

	// if last equals generated last with base and mask
	base := baseIP(a.base, mask)
	last := lastIP(a.base, mask)
	if base == a.base && last == a.last {
		return mask
	}
	// Block is no CIDR
	return ipZero
}

// BlockToCIDRList returns a list of CIDRs spanning the range of a.
func (a Block) BlockToCIDRList() []Block {
	if a.IsCIDR() {
		return []Block{a}
	}

	// here we go
	out := make([]Block, 0)

	// v4 or v6
	bits := 32
	if a.base.octets[0] == IPv6 {
		bits = 128
	}

	// start values
	cursor := a.base
	end := a.last

	// break condition: last == end, see below
	for {
		maskLen, hostLen := Block{base: cursor, last: end}.bitLen()
		mask := setBytes(net.CIDRMask(maskLen, bits))

		// find matching bitlen/mask at cursor position
		for hostLen > 0 {
			s := baseIP(cursor, mask) // make start with new mask
			l := lastIP(cursor, mask) // make last with new mask

			// bitlen is ok, if s = (cursor & mask) is still equal to cursor
			// and the new calculated last is still <= a.Last
			if s == cursor && l.octets <= end.octets {
				break
			}

			// nope, no success, reduce bitlen and try again
			hostLen--
			mask = setBytes(net.CIDRMask(bits-hostLen, bits))
		}

		base := baseIP(cursor, mask)
		last := lastIP(cursor, mask)
		cidr := Block{base: base, last: last, mask: mask}

		out = append(out, cidr)

		// stop if last == end
		if last == end {
			break
		}

		// move the cursor one behind the current last
		var ok bool
		if cursor, ok = last.addOne(); !ok {
			panic(errOverflow)
		}
	}

	return out
}

// Aggregate returns the minimal number of CIDRs spanning the range of input blocks.
func Aggregate(bs []Block) []Block {
	if len(bs) == 0 {
		return nil
	}

	// first step: expand input blocks (maybe ranges) to real CIDRs
	// use a set, get rid of dup cidrs
	set := map[Block]bool{}
	for i := range bs {
		cidrs := bs[i].BlockToCIDRList()
		for _, cidr := range cidrs {
			set[cidr] = true
		}
	}

	// next step: maybe we still have supersets and subsets, remove the subsets
	// back from map to slice
	cidrs := make([]Block, 0, len(set))
	for cidr := range set {
		cidrs = append(cidrs, cidr)
	}

	// sort slice
	sort.Slice(cidrs, func(i, j int) bool { return cidrs[i].Less(cidrs[j]) })

	// skip subsets
	unique := make([]Block, 0, len(cidrs))
	for i := 0; i < len(cidrs); i++ {
		unique = append(unique, cidrs[i])

		var cursor int
		for j := i + 1; j < len(cidrs); j++ {
			if cidrs[i].Covers(cidrs[j]) {
				cursor = j
				continue
			}
			break
		}

		// move cursor
		if cursor != 0 {
			i = cursor
		}
	}

	// next step: no more subsets, pack adjacent cidrs to blocks
	packed := make([]Block, 0, len(unique))

	for i := 0; i < len(unique); i++ {
		pack := unique[i]
		var cursor int

		// pack adjacencies to cursor at pos i
		for j := i + 1; j < len(unique); j++ {

			// break on end of IP address space, else panic on overflow!
			if pack.last == ipMaxV4 || pack.last == ipMaxV6 {
				break
			}

			// test for adjacency, no gap between two cidrs
			look, _ := pack.last.addOne()
			if look == unique[j].base {
				// combine adjacent cidrs
				pack.last = unique[j].last
				cursor = j
				continue
			}
			break
		}

		// we changed pack.Last, calculate new pack.Mask, maybe it's no CIDR, just a range now
		pack.mask = pack.getMask()
		packed = append(packed, pack)

		// move cursor
		if cursor != 0 {
			i = cursor
		}
	}

	// last step: expand packed blocks (maybe ranges) to real CIDRs
	out := make([]Block, 0, len(packed))
	for _, r := range packed {
		cidrList := r.BlockToCIDRList()
		out = append(out, cidrList...)
	}

	return out
}
