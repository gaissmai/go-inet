package inet

import "errors"

type (
	// IP represents a single IPv4 or IPv6 address in a fixed array of 21 bytes.
	//
	//  IP[0]    = version information (4 or 6)
	//  IP[1:5]  = IPv4 address, if version == 4, else zero
	//  IP[5:21] = IPv6 address, if version == 6, else zero
	//
	// This IP representation is comparable and can be used as key in maps
	// and fast sorted by bytes.Compare() without conversions to/from the different IP versions.
	IP [21]byte

	// Block is an IP-network or IP-range, e.g.
	//
	//  192.168.0.1/24              // network, with CIDR mask
	//  ::1/128                     // network, with CIDR mask
	//  10.0.0.3-10.0.17.134        // range
	//  2001:db8::1-2001:db8::f6    // range
	//
	// This Block representation is comparable and can be used as key in maps
	// and fast sorted without conversions to/from the different IP versions.
	Block struct {
		Base IP
		Last IP
		Mask IP // IPZero for ranges without CIDR mask
	}

	// Tree is an implementation of a CIDR/Block tree for fast IP lookup with longest-prefix-match.
	// It is NOT a standard patricia-trie, this isn't possible for general blocks not represented by bitmasks.
	Tree struct {
		// Contains the root node of a tree.
		Root *Node
	}

	// Node, recursive data structure. Items are abstracted via Itemer interface
	Node struct {
		Item   *Itemer
		Parent *Node
		Childs []*Node
	}

	// Itemer interface for tree items, maybe with payload and not just ip Blocks.
	// See relation between Compare and Contains at inet.Block.Compare()
	Itemer interface {

		// Contains, defines the depth in the tree, parent child relationship.
		Contains(Itemer) bool

		// Compare, defines equality and sort order on same tree level, siblings relationship.
		Compare(Itemer) int
	}
)

var (
	// IPZero is the zero-value for type IP.
	//
	// IP is represented as an array, so we have no nil as zero value.
	// IPZero can be used for that.
	IPZero = IP{}

	// BlockZero is the zero-value for type Block.
	//
	// Block is represented as a struct, so we have no nil as zero value.
	// BlockZero can be used for that.
	BlockZero = Block{Base: IP{}, Last: IP{}, Mask: IP{}}

	// MaxCIDRSplit limits the input parameter for SplitCIDR() to 20 (max 2^20 CIDRs) for security.
	MaxCIDRSplit int = 20

	// internal for overflow checks in aggregates
	ipMaxV4 = IP{4, 255, 255, 255, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	ipMaxV6 = IP{6, 0, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

var (
	ErrInvalidIP       = errors.New("invalid IP")
	ErrInvalidBlock    = errors.New("invalid Block")
	ErrVersionMismatch = errors.New("IPv4/IPv6 version mismatch")
	ErrOverflow        = errors.New("overflow")
	ErrUnderflow       = errors.New("underflow")
)
