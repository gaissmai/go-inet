package inet

import (
	"bytes"
	"fmt"
	"net"
)

// ########################################################
// implementations for type IP
// ########################################################

// String implements the fmt.Stringer interface.
// Returns "" on IPZero, panics otherwise on invalid input.
func (ip IP) String() string {
	if ip == IPZero {
		return ""
	}

	if !ip.IsValid() {
		panic(ErrInvalidIP)
	}

	return ip.ToNetIP().String()
}

// MarshalText implements the encoding.TextMarshaler interface.
// The encoding is the same as returned by String.
func (ip IP) MarshalText() ([]byte, error) {
	return []byte(ip.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The IP address is expected in a form accepted by ParseIP(string).
func (ip *IP) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) == 0 { // this is no error condition
		*ip = IPZero
		return nil
	}

	x, err := ipFromString(s)
	if err != nil {
		return err
	}

	*ip = x
	return nil
}

// ########################################################
// implementations for type Block
// ########################################################

// String implements the fmt.Stringer interface.
// Returns "" on BockZero.
func (a Block) String() string {
	if a == BlockZero {
		return ""
	}

	if !a.IsValid() {
		panic(ErrInvalidBlock)
	}

	if a.Mask == IPZero {
		return fmt.Sprintf("%s-%s", a.Base, a.Last)
	}

	mb := a.Mask.Bytes()
	ones, _ := net.IPMask(mb).Size()
	return fmt.Sprintf("%s/%d", a.Base, ones)
}

// MarshalText implements the encoding.TextMarshaler interface.
// The encoding is the same as returned by String.
func (a Block) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// The block is expected in a form accepted by ParseBlock.
func (a *Block) UnmarshalText(text []byte) error {
	s := string(text)
	if len(s) == 0 { // this is no error condition
		*a = BlockZero
		return nil
	}

	x, err := blockFromString(s)
	if err != nil {
		return err
	}

	*a = x
	return nil
}

// Contains reports whether Block a contains Block b. a and b may NOT coincide.
// Implements the tree.Itemer interface.
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
// Implements the tree.Itemer interface.
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
