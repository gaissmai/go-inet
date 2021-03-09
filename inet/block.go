package inet

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
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
//
// Each Block object only stores two IP addresses, the base and last address of the range ot network.
type Block struct {
	base IP
	last IP
}

var (
	errInvalidBlock = errors.New("invalid Block")
	errOverflow     = errors.New("overflow")
	errUnderflow    = errors.New("underflow")
)

// the zero-value of type Block
var blockZero Block

// Base returns the blocks base IP address.
func (b Block) Base() IP { return b.base }

// Last returns the blocks last IP address.
func (b Block) Last() IP { return b.last }

// ParseBlock parses and returns the input as type Block.
// The input type may be:
//
//   string
//   net.IPNet
//   net.IP
//   IP
//
// Example for valid input strings:
//
//  "192.168.2.3-192.168.7.255"
//  "2001:db8::1-2001:db8::ff00:35"
//
//  "2001:db8:dead::/38"
//  "10.0.0.0/8"
//  "4.4.4.4"
//
// IP addresses as input are converted to /32 or /128 blocks
// Returns error and Block{} on invalid input.
//
// The hard part is done by net.ParseIP() and net.ParseCIDR().
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

// blockFromIP converts inet.IP to inet.Block with ip as base and last.
func blockFromIP(ip IP) (Block, error) {
	b := Block{base: ip, last: ip}
	return b, nil
}

// blockFromNetIP converts net.IP to inet.Block with /32 or /128 CIDR mask
func blockFromNetIP(stdIP net.IP) (Block, error) {
	ip, err := fromStdIP(stdIP)
	if err != nil {
		return blockZero, err
	}
	return blockFromIP(ip)
}

// blockFromNetIPNet converts from stdlib net.IPNet to ip.Block representation.
func blockFromNetIPNet(stdNet net.IPNet) (Block, error) {
	var err error
	a := blockZero

	a.base, err = fromStdIP(stdNet.IP)
	if err != nil {
		return blockZero, errInvalidBlock
	}

	mask, err := fromStdIP(net.IP(stdNet.Mask)) // cast needed
	if err != nil {
		return blockZero, errInvalidBlock
	}

	a.last = a.base.mkLastIP(mask.uint128)

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
	if baseIP.version != lastIP.version {
		return blockZero, errInvalidBlock
	}

	// begin > end
	if !baseIP.Less(lastIP) {
		return blockZero, errInvalidBlock
	}

	return Block{base: baseIP, last: lastIP}, nil
}

// IsValid reports whether block is valid and not the zero value of the Block type.
// The zero value is not a valid Block of any type.
func (b Block) IsValid() bool {
	return b.base.IsValid() && b.last.IsValid()
}

// Is4 reports whether block is IPv4.
func (b Block) Is4() bool {
	return b.base.version == v4
}

// Is6 reports whether block is IPv6.
func (b Block) Is6() bool {
	return b.base.version == v6
}

// IsCIDR returns true if the block has a common prefix netmask.
func (b Block) IsCIDR() bool {
	return b.base.isCIDR(b.last)
}

// String returns the string form of the Block.
// It returns one of 3 forms:
//
//   - "invalid Block", if IsValid is false
//   - as range: "127.0.0.1-127.0.0.19", if block is no CIDR
//   - as CIDR:  "2001:db8::/32"
func (b Block) String() string {
	if b == blockZero {
		return "invalid Block"
	}
	if !b.IsCIDR() {
		return fmt.Sprintf("%s-%s", b.base, b.last)
	}

	n := b.base.commonPrefixLen(b.last)
	if b.base.version == v4 {
		n = n - 96
	}
	return fmt.Sprintf("%s/%d", b.base, n)
}

// Covers reports whether Block a contains Block b. a and b may NOT coincide.
// a.Covers(b) returns true when a is a *true* cover of b, a == b must then be false.
//
//  a |-----------------| |-----------------| |-----------------|
//  b   |------------|    |------------|           |------------|
func (b Block) Covers(c Block) bool {
	if b.base.version != c.base.version {
		return false
	}
	if b == c {
		return false
	}
	if b.base.uint128.cmp(c.base.uint128) <= 0 && b.last.uint128.cmp(c.last.uint128) >= 0 {
		return true
	}
	return false
}

// Less reports whether the a should be sorted before b.
// REMEMBER: sort supersets always to the left of their subsets!
// If b.Covers(c) is true then b.Less(c) must also be true.
//
//  b |---|
//  c       |------|
//
//  b |-------|
//  c    |------------|
//
//  b |-----------------|
//  c    |----------|
//
//  b |-----------------|
//  c |------------|
func (b Block) Less(c Block) bool {
	if b.base.Less(c.base) {
		return true
	}

	if b.base == c.base { // ... and a covers b,
		//	REMEMBER: sort containers to the left
		return b.last.uint128.cmp(c.last.uint128) == 1
	}

	return false
}

// Merge adjacent blocks, remove dups and subsets, returns the remaining blocks sorted.
func Merge(bs []Block) []Block {
	switch len(bs) {
	case 0:
		return nil
	case 1:
		return []Block{bs[0]}
	}

	// must be sorted for this algo!
	sort.Slice(bs, func(i, j int) bool { return bs[i].Less(bs[j]) })

	out := make([]Block, 1, len(bs))
	out[0] = bs[0]
	for _, b := range bs[1:] {
		prev := &out[len(out)-1]
		switch {
		case b == blockZero:
			// no-op
		case prev.overlaps(b):
			prev.last = b.last
		case prev.last.addOne() == b.base:
			prev.last = b.last
		case prev.isDisjunct(b):
			out = append(out, b)
		default:
			// no-op: covers or equal
		}
	}
	return out
}

// CIDRs returns a list of CIDRs that span b.
func (b Block) CIDRs() []Block {
	if b == blockZero {
		return nil
	}
	return b.base.toCIDRsRec(nil, b.last)
}

// recursion ahead
// end condition: isCIDR
// split the range in the middle
// call both halves recursively
func (a IP) toCIDRsRec(buf []Block, b IP) []Block {
	if a.isCIDR(b) {
		buf = append(buf, Block{a, b})
		return buf
	}

	// get next mask (+1 Bit)
	n := a.commonPrefixLen(b)
	m := maskUint128[n+1]

	// split range with new mask s, s+1
	u := a.mkLastIP(m)
	v := b.mkBaseIP(m)

	// rec call for both halves, {a, u} and {v, b}
	buf = a.toCIDRsRec(buf, u)
	buf = v.toCIDRsRec(buf, b)

	return buf
}

// Diff the slice of blocks from receiver, returns the remaining blocks.
func (b Block) Diff(bs []Block) []Block {
	// nothing to remove
	if len(bs) == 0 {
		return []Block{b}
	}

	// to remove blocks must be sorted for this algo!
	sort.Slice(bs, func(i, j int) bool { return bs[i].Less(bs[j]) })

	var out []Block
	for _, d := range bs {
		switch {
		case d == blockZero:
			// no-op
		case d.isDisjunct(b):
			// no-op
		case d == b:
			// masks rest
			return out
		case d.Covers(b):
			// masks rest
			return out
		case d.base == b.base:
			// move forward
			b.base = d.last.addOne()
		case b.base.Less(d.base):
			// save [b.base, d.base)
			out = append(out, Block{b.base, d.base.subOne()})
			// new b, (d.last, b.last]
			b.base = d.last.addOne()
		default:
			panic("logic error")
		}
		// overflow from last addOne()
		if b.base == ipZero {
			return out
		}
		// cursor moved behind b.last
		if b.last.Less(b.base) {
			return out
		}
	}
	// save the rest
	out = append(out, b)

	return out
}

// isDisjunct reports whether the Blocks b and c are disjunct
//  b       |----------|
//  c |---|
//
//  b |------|
//  c          |---|
func (b Block) isDisjunct(c Block) bool {

	//  a       |----------|
	//  b |---|
	if c.last.Less(b.base) {
		return true
	}

	//  a |------|
	//  b          |---|
	if b.last.Less(c.base) {
		return true
	}

	return false
}

// overlaps reports whether the Blocks overlaps.
//
//  b    |-------|
//  c |------|
//
//  b |------|
//  c    |-------|
//
//  b |----|
//  c      |---------|
//
//  b      |---------|
//  c |----|
func (b Block) overlaps(c Block) bool {
	if b == c {
		return false
	}
	if b.Covers(c) || c.Covers(b) {
		return false
	}
	if b.isDisjunct(c) {
		return false
	}
	return true
}
