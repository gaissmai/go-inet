package tree

import (
	"fmt"
	"net"

	"github.com/gaissmai/go-inet/inet"
)

// Itemer interface for tree items, maybe with payload, not just ip Blocks.
type Itemer interface {

	// Contains, defines the depth in the tree, parent child relationship.
	Contains(Itemer) bool

	// Compare, defines equality and sort order on same tree level, siblings relationship.
	Compare(Itemer) int
}

// simpleItem is just a wrapper for inet.Block, implementing the Itemer interface.
// Used for building BlockTrees, no other payload like text or next-hop.
type simpleItem struct{ inet.Block }

// Contains, part of Itemer interface
func (a simpleItem) Contains(b Itemer) bool {
	if other, ok := b.(simpleItem); ok {
		return a.Block.Contains(other.Block)
	}
	panic(fmt.Errorf("incompatible types: %T != %T", a, b))
}

// Compare, part of Itemer interface, compares just the blocks.
func (a simpleItem) Compare(b Itemer) int {
	if other, ok := b.(simpleItem); ok {
		return a.Block.Compare(other.Block)
	}
	panic(fmt.Errorf("incompatible types: %T != %T", a, b))
}

// NewSimpleItem wraps the input as type Itemer.
//
// The input type may be:
//   inet.IP
//   inet.Block
//   net.IP
//   net.IPNet
//  *net.IP
//  *net.IPNet
//
// Plain addresses are converted to host blocks, .../32 for IPv4 and .../128 for IPv6.
func NewSimpleItem(i interface{}) Itemer {
	switch i := i.(type) {
	case inet.IP:
		b := inet.MustBlock(inet.NewBlock(i.String() + "-" + i.String()))
		return simpleItem{b}
	case net.IP:
		b := inet.MustBlock(inet.NewBlock(i.String() + "-" + i.String()))
		return simpleItem{b}
	case *net.IP:
		b := inet.MustBlock(inet.NewBlock((*i).String() + "-" + (*i).String()))
		return simpleItem{b}
	case net.IPNet, *net.IPNet:
		b := inet.MustBlock(inet.NewBlock(i))
		return simpleItem{b}
	case inet.Block:
		return simpleItem{i}
	default:
		panic(fmt.Errorf("unexpected type %t", i))
	}
}
