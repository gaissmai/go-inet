package inet

import (
	"bytes"
	"errors"
	"math/big"
	"net"
	"sort"
	"strings"
)

var (
	errInvalidBlock = errors.New("invalid Block")
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
	Base IP
	Last IP
	Mask IP // IP{} for ranges without CIDR mask
}

// the zero-value for type Block, not public
var blockZero Block = Block{}

var (
	// MaxCIDRSplit limits the input parameter for SplitCIDR() to 20 (max 2^20 CIDRs) for security.
	MaxCIDRSplit int = 20

	// internal for overflow checks in aggregates
	ipMaxV4 = IP{4, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	ipMaxV6 = IP{6, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
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
//  "2001:db8:dead:/38"
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
	ip, err := ipFromString(s)
	if err == nil {
		return blockFromIP(ip)
	}

	return blockZero, errInvalidBlock
}

// blockFromIP converts inet.IP to inet.Block with /32 or /128 CIDR mask
func blockFromIP(ip IP) (Block, error) {
	b := Block{Base: ip, Last: ip}
	b.Mask = b.getMask()
	return b, nil
}

// blockFromNetIP converts net.IP to inet.Block with /32 or /128 CIDR mask
func blockFromNetIP(nip net.IP) (Block, error) {
	ip, err := ipFromNetIP(nip)
	if err != nil {
		return blockZero, err
	}
	return blockFromIP(ip)
}

// blockFromNetIPNet converts from stdlib net.IPNet to ip.Block representation.
func blockFromNetIPNet(ipnet net.IPNet) (Block, error) {
	var err error
	a := blockZero

	a.Base, err = ipFromNetIP(ipnet.IP)
	if err != nil {
		return blockZero, errInvalidBlock
	}

	a.Mask, err = ipFromNetIP(net.IP(ipnet.Mask)) // cast needed
	if err != nil {
		return blockZero, errInvalidBlock
	}

	a.Last = lastIP(a.Base, a.Mask)

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
	if baseIP.Version() != lastIP.Version() {
		return blockZero, errInvalidBlock
	}

	// begin > end
	if baseIP.Compare(lastIP) == 1 {
		return blockZero, errInvalidBlock
	}

	a := Block{Base: baseIP, Last: lastIP}

	// maybe range is true CIDR, try to get also the mask
	a.Mask = a.getMask()

	return a, nil
}

// Contains reports whether Block a contains Block b. a and b may NOT coincide.
//
//  a   |------------|    |------------|           |------------|
//  b |-----------------| |-----------------| |-----------------|
func (a Block) Contains(b Block) bool {
	if a == b {
		return false
	}
	return bytes.Compare(a.Base[:], b.Base[:]) <= 0 && bytes.Compare(a.Last[:], b.Last[:]) >= 0
}

// Compare returns an integer comparing two IP Blocks.
//
//   0 if a == b,
//
//  -1 if a is v4 and b is v6
//  +1 if a is v6 and b is v4
//
//  -1 if a.Base < b.Base
//  +1 if a.Base > b.Base
//
//  -1 if a.Base == b.Base and a is SuperSet of b
//  +1 if a.Base == b.Base and a is Subset of b
func (a Block) Compare(b Block) int {
	if bytes.Compare(a.Base[:], b.Base[:]) < 0 {
		return -1
	}
	if bytes.Compare(a.Base[:], b.Base[:]) > 0 {
		return 1
	}
	// base is now equal, test for superset/subset
	if bytes.Compare(a.Last[:], b.Last[:]) > 0 {
		return -1
	}
	if bytes.Compare(a.Last[:], b.Last[:]) < 0 {
		return 1
	}
	return 0
}

// IsCIDR returns true if the block is a true CIDR, not just a begin-end range.
func (a Block) IsCIDR() bool {
	return a.Mask != ipZero
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
		panic(errInvalidBlock)
	}
	return a.Base.Version()
}

// BitLen returns the minimum number of bits to represent the block.
func (a Block) BitLen() int {
	// algorithm: use math.big.BitLen(lastIP-baseIP)
	ip := a.Last
	ip = ip.SubBytes(a.Base.Bytes())
	return new(big.Int).SetBytes(ip.Bytes()).BitLen()
}

// Size returns the number of ip addresses as string.
// Returns a string, since the amount of ip addresses can be greater than uint64.
func (a Block) Size() string {
	// algorithm: use math.big.String(lastIP-baseIP+1)
	ip := a.Last
	diff := new(big.Int).SetBytes(ip.SubBytes(a.Base.Bytes()).Bytes())
	one := new(big.Int).SetInt64(1)
	return diff.Add(diff, one).String()
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
	// - make next base and next last with new mask
	// - break if next last == a.last
	// - ... or increment next last, use it as new base
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
		next := blockZero
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

// FindFreeCIDR returns all free CIDR blocks (of max possible bitlen) within given CIDR,
// minus the inner CIDR blocks.
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

	SortBlock(free)
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
		if b.Contains(a) {
			return true
		}
	}
	return false
}

// getMask is a helper method. Calculate a netmask from begin-end, returns IP{} if not possible.
func (a Block) getMask() IP {
	// v4 or v6
	bits := 32
	if a.Base.Version() == 6 {
		bits = 128
	}

	// bits for hostmask
	bitlen := a.BitLen()

	// netmask is inverse of hostmask, bits-bitlen
	mask := setBytes(net.CIDRMask(bits-bitlen, bits))

	// if last equals generated last with base and mask
	base := baseIP(a.Base, mask)
	last := lastIP(a.Base, mask)
	if base == a.Base && last == a.Last {
		return mask
	}
	return ipZero
}

// BlockToCIDRList returns a list of CIDRs spanning the range of a.
func (a Block) BlockToCIDRList() []Block {
	if !a.IsValid() {
		panic(errInvalidBlock)
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
		bitlen := Block{Base: cursor, Last: end}.BitLen()
		mask := setBytes(net.CIDRMask(bits-bitlen, bits))

		// find matching bitlen/mask at cursor position
		for bitlen > 0 {
			s := baseIP(cursor, mask) // try start
			l := lastIP(cursor, mask) // try last

			// bitlen is ok, if s = (cursor & mask) is still equal to cursor
			// and the new calculated last is still <= a.Last
			if s.Compare(cursor) == 0 && l.Compare(end) <= 0 {
				break
			}

			// nope, no success, reduce bitlen and try again
			bitlen--
			mask = setBytes(net.CIDRMask(bits-bitlen, bits))
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
	SortBlock(cidrs)

	// skip subsets
	unique := make([]Block, 0, len(cidrs))
	for i := 0; i < len(cidrs); i++ {
		unique = append(unique, cidrs[i])

		var cursor int
		for j := i + 1; j < len(cidrs); j++ {
			if cidrs[i].Contains(cidrs[j]) {
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
			if pack.Last == ipMaxV4 || pack.Last == ipMaxV6 {
				break
			}

			// test for adjacency, no gap between two cidrs
			if pack.Last.AddUint64(1) == unique[j].Base {
				// combine adjacent cidrs
				pack.Last = unique[j].Last
				cursor = j
				continue
			}
			break
		}

		// we changed pack.Last, calculate new pack.Mask, maybe it's no CIDR, just a range now
		pack.Mask = pack.getMask()
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
