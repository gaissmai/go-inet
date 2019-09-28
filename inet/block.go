package inet

import (
	"bytes"
	"math/big"
	"net"
	"sort"
	"strings"
)

// NewBlock parses and returns the input as type Block.
// The input type may be:
//   net.IPNet
//  *net.IPNet
//   string
// If a begin-end range can be represented as a CIDR, NewBlock() generates the netmask and returns the CIDR.
//
// IP addresses are converted to /32 oder /128 blocks
//  inet.IP
//   net.IP
//  *net.IP
//
// The hard part parsing string representation is done by net.ParseCIDR().
// Returns error and BlockZero on invalid input.
//
func NewBlock(i interface{}) (Block, error) {
	switch v := i.(type) {
	case string:
		return blockFromString(v)
	case net.IPNet:
		return blockFromNetIPNet(v)
	case *net.IPNet:
		return blockFromNetIPNet(*v)
	case IP:
		b := Block{Base: v, Last: v}
		b.Mask = b.getMask()
		return b, nil
	case net.IP:
		ip, err := ipFromNetIP(v)
		if err != nil {
			return BlockZero, err
		}
		b := Block{Base: ip, Last: ip}
		b.Mask = b.getMask()
		return b, nil
	case *net.IP:
		ip, err := ipFromNetIP(*v)
		if err != nil {
			return BlockZero, err
		}
		b := Block{Base: ip, Last: ip}
		b.Mask = b.getMask()
		return b, nil
	default:
		return BlockZero, ErrInvalidBlock
	}
}

// MustBlock is a helper that calls NewBlock and returns just inet.Block or panics on errr.
// It is intended for use in variable initializations.
func MustBlock(i interface{}) Block {
	b, err := NewBlock(i)
	if err != nil {
		panic(err)
	}
	return b
}

// blockFromNetIPNet converts from stdlib net.IPNet to ip.Block representation.
func blockFromNetIPNet(ipnet net.IPNet) (Block, error) {
	var err error
	var a = BlockZero

	a.Base, err = ipFromNetIP(ipnet.IP)
	if err != nil {
		return BlockZero, ErrInvalidBlock
	}

	a.Mask, err = ipFromNetIP(net.IP(ipnet.Mask)) // cast needed
	if err != nil {
		return BlockZero, ErrInvalidBlock
	}

	a.Last = lastIP(a.Base, a.Mask)

	return a, nil
}

// blockFromString parses s in network CIDR or in begin-end IP address-range notation.
func blockFromString(s string) (Block, error) {
	if s == "" {
		return BlockZero, ErrInvalidBlock
	}

	i := strings.IndexByte(s, '/')
	if i >= 0 {
		return newBlockFromCIDR(s)
	}

	i = strings.IndexByte(s, '-')
	if i >= 0 {
		return newBlockFromRange(s, i)
	}

	return BlockZero, ErrInvalidBlock
}

// parse IP CIDR
// e.g.: 127.0.0.0/8 or 2001:db8::/32
func newBlockFromCIDR(s string) (Block, error) {
	_, netIPNet, err := net.ParseCIDR(s)
	if err != nil {
		return BlockZero, err
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
	b := base.Bytes()
	m := mask.Bytes()

	last := make([]byte, len(b))
	for i := 0; i < len(b); i++ {
		last[i] = b[i] | ^m[i]
	}
	return setBytes(last)
}

// baseIP makes base address from address and netmask:
//  base[i] = address[i] & netMask[i]
func baseIP(any IP, mask IP) IP {
	a := any.Bytes()
	m := mask.Bytes()

	base := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		base[i] = a[i] & m[i]
	}
	return setBytes(base)
}

// parse IP address-range
// e.g.: 127.0.0.0-127.0..0.17 or 2001:db8::1-2001:dbb::ffff
func newBlockFromRange(s string, i int) (Block, error) {
	// split string
	base, last := s[:i], s[i+1:]

	baseIP, err := NewIP(base)
	if err != nil {
		return BlockZero, ErrInvalidBlock
	}

	lastIP, err := NewIP(last)
	if err != nil {
		return BlockZero, ErrInvalidBlock
	}

	// begin-end have version mismatch
	if baseIP.Version() != lastIP.Version() {
		return BlockZero, ErrInvalidBlock
	}

	// begin > end
	if baseIP.Compare(lastIP) == 1 {
		return BlockZero, ErrInvalidBlock
	}

	a := Block{Base: baseIP, Last: lastIP}

	// maybe range is true CIDR, try to get also the mask
	a.Mask = a.getMask()

	return a, nil
}

// IsCIDR returns true if the block is a true CIDR, not just a begin-end range.
func (a Block) IsCIDR() bool {
	return a.Mask != IPZero
}

// IsValid returns true on valid Blocks, false otherwise.
func (a Block) IsValid() bool {
	if !a.Base.IsValid() || !a.Last.IsValid() {
		return false
	}

	// version mismatch
	if a.Base.Version() != a.Last.Version() {
		return false
	}

	// base is greater than last
	if a.Base.Compare(a.Last) > 0 {
		return false
	}

	// calc mask from base-last, compare with a.Mask
	m := a.getMask()

	// Mask is just an IP, IPs are comparable
	return m == a.Mask
}

// Version returns the IP version, 4 or 6, panics on invalid block.
func (a Block) Version() int {
	if !a.IsValid() {
		panic(ErrInvalidBlock)
	}
	return a.Base.Version()
}

// Size returns the minimum size in bits to represent the block.
func (a Block) Size() int {
	// algorithm: use math.big.BitLen(lastIP-baseIP)
	ip := a.Last
	ip = ip.SubBytes(a.Base.Bytes())
	return new(big.Int).SetBytes(ip.Bytes()).BitLen()
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
	if bytes.Compare(a.Base[:], b.Last[:]) == 1 {
		return true
	}

	//  a |------|
	//  b          |---|
	if bytes.Compare(a.Last[:], b.Base[:]) == -1 {
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
	if a.Contains(b) || b.Contains(a) {
		return false
	}
	if a.IsDisjunctWith(b) {
		return false
	}
	return true
}

// SortBlock sorts the given slice of Blocks in place, see Compare() for sort order.
// IPv4 Blocks are sorted before IPv6 Blocks. Outer sets are sorted before their subsets!
func SortBlock(bs []Block) {
	sort.Slice(bs, func(i, j int) bool { return bs[i].Compare(bs[j]) == -1 })
}

// SplitCIDR returns the next 2^n CIDRs, splitted from outer block.
// The number of CIDRs is limited to MaxCIDRSplit, panics if more CIDRs are requested.
// Returns nil at max mask length or if block is no CIDR.
func (a Block) SplitCIDR(n int) []Block {
	// algorithm:
	// - create new mask
	// loop
	// - make base and last with new mask
	// - break if new last == a.mask
	// - increment last, use it as next base
	// end

	// limit cpu and memory
	if n > MaxCIDRSplit {
		panic("too many CIDRs requested")
	}

	if !a.IsCIDR() {
		return nil
	}

	// check for max mask len, bits are 32 or 128 (v4 or v6)
	maskSize, bits := net.IPMask(a.Mask.Bytes()).Size()
	if n <= 0 || maskSize+n > bits {
		return nil
	}

	newMask := setBytes(net.CIDRMask(maskSize+n, bits))

	cidrs := make([]Block, 0, 1<<uint(n))

	base := a.Base
	if baseIP(base, newMask) != base {
		panic("logic error ...")
	}

	for {
		next := BlockZero
		next.Base = baseIP(base, newMask)
		next.Last = lastIP(next.Base, newMask)
		next.Mask = newMask
		cidrs = append(cidrs, next)

		// last of outer CIDR already reached?
		if next.Last.Compare(a.Last) == -1 {
			base = next.Last.AddUint64(1) // next base
		} else if next.Last == a.Last {
			break
		} else {
			panic("logic error...")
		}
	}
	return cidrs
}

// FindOuterCIDR returns the next enclosing CIDR (SuperSet) for all the inner CIDRs or ranges together.
// All blocks must have the same IP version, panics on invalid input.
func FindOuterCIDR(bs []Block) Block {
	if len(bs) == 0 {
		return BlockZero
	}

	// all blocks must have the same IP version
	version := bs[0].Version()
	for _, b := range bs {
		if b.Version() != version {
			panic(ErrVersionMismatch)
		}
	}

	SortBlock(bs)

	bits := 32
	if version == 6 {
		bits = 128
	}

	lower := bs[0].Base
	upper := bs[len(bs)-1].Last

	hostSize := Block{Base: lower, Last: upper}.Size()
	maskSize := bits - hostSize

	mask := setBytes(net.CIDRMask(maskSize, bits))
	base := baseIP(lower, mask)
	last := lastIP(base, mask)

	return Block{Base: base, Last: last, Mask: mask}
}

// FindFreeCIDR returns all free CIDR blocks (of max possible size) within given CIDR, minus the inner CIDR blocks.
// Panics if inner blocks are no subset of (or not equal to) outer block.
func (a Block) FindFreeCIDR(bs []Block) []Block {
	for _, i := range bs {
		if !(a.Contains(i) || i == a) {
			panic("at least one inner block isn't contained in (or equal to) outer block")
		}
	}

	free := make([]Block, 0, 10)

	candidates := make([]Block, 0, 10)
	candidates = append(candidates, a) // start with outer block

	for i := 0; i < len(candidates); i++ {
		c := candidates[i]

		if c == BlockZero {
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

	SortBlock(free)
	return free
}

// isDisjunctWithAll is a helper method. Returns true if b is disjunct with any inner block
func (a Block) isDisjunctWithAll(bs []Block) bool {
	for _, b := range bs {
		if !a.IsDisjunctWith(b) {
			return false
		}
	}
	return true
}

// isSubsetOfAny is a helper method: Returns true if b is subset of any inner block
func (a Block) isSubsetOfAny(bs []Block) bool {
	for _, b := range bs {
		if b.Contains(a) {
			return true
		}
	}
	return false
}

// getMask is a helper method. Calculate a netmask from begin-end, returns IPZero if not possible.
func (a Block) getMask() IP {
	// v4 or v6
	bits := 32
	if a.Base.Version() == 6 {
		bits = 128
	}

	// bits for hostmask
	size := a.Size()

	// netmask is inverse of hostmask, bits-size
	mask := setBytes(net.CIDRMask(bits-size, bits))

	// if last equals generated last with base and mask
	base := baseIP(a.Base, mask)
	last := lastIP(a.Base, mask)
	if base == a.Base && last == a.Last {
		return mask
	}
	return IPZero
}

// BlockToCIDRList returns a list of CIDRs spanning the range of a.
func (a Block) BlockToCIDRList() []Block {
	if !a.IsValid() {
		panic(ErrInvalidBlock)
	}

	if a.IsCIDR() {
		return []Block{a}
	}

	// here we go
	out := make([]Block, 0)

	// v4 or v6
	bits := 32
	if a.Base.Version() == 6 {
		bits = 128
	}

	// start values for loop
	cursor := a.Base
	end := a.Last

	// stop condition, cursor > end
	for cursor.Compare(end) <= 0 {
		size := Block{Base: cursor, Last: end}.Size()
		mask := setBytes(net.CIDRMask(bits-size, bits))

		// find matching size/mask at cursor position
		for size > 0 {
			s := baseIP(cursor, mask) // try start
			l := lastIP(cursor, mask) // try last

			// size is ok, if s = (cursor & mask) is still equal to cursor
			// and the new calculated last is still <= a.Last
			if s.Compare(cursor) == 0 && l.Compare(end) <= 0 {
				break
			}

			// nope, no success, reduce size and try again
			size--
			mask = setBytes(net.CIDRMask(bits-size, bits))
		}

		base := baseIP(cursor, mask)
		last := lastIP(cursor, mask)
		cidr := Block{Base: base, Last: last, Mask: mask}

		out = append(out, cidr)

		// move the cursor one behind last
		cursor = last.AddUint64(1)
	}

	return out
}

// Aggregate, returns the minimal number of CIDRs spanning the range of input blocks.
func Aggregate(bs []Block) []Block {
	if len(bs) == 0 {
		return nil
	}

	// first step: expand input blocks (maybe ranges) to real CIDRs
	cidrs := make([]Block, 0, len(bs))
	for i := range bs {
		expand := bs[i].BlockToCIDRList()
		cidrs = append(cidrs, expand...)
	}

	// next step: real cidrs have no overlaps, just dups and subsets
	// remove the dups and subsets, keep just unique cidrs
	unique := make([]Block, 0, len(cidrs))

	// use sort order
	SortBlock(cidrs)

	for i := 0; i < len(cidrs); i++ {
		unique = append(unique, cidrs[i])
		skip := 0

		// skip dups and subsets after cursor at pos i
		for j := i + 1; j < len(cidrs); j++ {

			// skip dups
			if cidrs[i].Compare(cidrs[j]) == 0 {
				// skip next identical cidr in row
				skip = j
				continue
			}

			// skip subsets
			if cidrs[i].Contains(cidrs[j]) {
				// skip next subset in row
				skip = j
				continue
			}
			break
		}

		// move cursor
		if skip != 0 {
			i = skip
		}
	}

	// next step: no more dups and subsets, pack adjacent cidrs to blocks
	packed := make([]Block, 0, len(unique))

	for i := 0; i < len(unique); i++ {
		pack := unique[i]
		skip := 0

		// pack adjacencies to cursor at pos i
		for j := i + 1; j < len(unique); j++ {

			// break on end of IP address space, else panic on overflow!
			if pack.Last == ipMaxV4 || pack.Last == ipMaxV6 {
				break
			}

			// test for adjacency, no gap between two cidrs
			if pack.Last.AddUint64(1) == unique[j].Base {
				// combine adjacent cidrs
				pack.Last = unique[j].Last
				skip = j
				continue
			}
			break
		}

		// we changed pack.Last, calculate new pack.Mask, maybe it's no CIDR, just a range now
		pack.Mask = pack.getMask()
		packed = append(packed, pack)

		// move cursor
		if skip != 0 {
			i = skip
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
