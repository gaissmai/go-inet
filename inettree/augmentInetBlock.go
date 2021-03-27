// Package inettree implements the tree.Interface for inet.Block
package inettree

import (
	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/tree"
)

// compiler check, Item implements tree.Interface
var _ tree.Interface = Item{}

// Item augments inet.Block, implementing the tree.Interface
type Item struct {
	// the augmented Block
	inet.Block

	// augment Block with additional text, see example
	Text string
}

// Less implements the tree.Interface for Item
func (a Item) Less(i tree.Interface) bool {
	b := i.(Item)
	return a.Block.Less(b.Block)
}

// Equal implements the tree.Interface for Item
func (a Item) Equals(i tree.Interface) bool {
	b := i.(Item)
	return a.Block == b.Block
}

// Covers implements the tree.Interface for Item
func (a Item) Covers(i tree.Interface) bool {
	b := i.(Item)
	return a.Block.Covers(b.Block)
}

// String implements the tree.Interface for Item
func (a Item) String() string {
	if a.Text == "" {
		// print just the Block as string
		return a.Block.String()
	}
	// augment the Block with additonal text
	return a.Text
}
