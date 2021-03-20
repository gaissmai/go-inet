package inet

import (
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
		panic(errInvalidIP)
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
// Returns "" on Block{}
func (a Block) String() string {
	if a == blockZero {
		return ""
	}

	if !a.IsValid() {
		panic(errInvalidBlock)
	}

	if a.Mask == "" {
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
		*a = blockZero
		return nil
	}

	x, err := blockFromString(s)
	if err != nil {
		return err
	}

	*a = x
	return nil
}
