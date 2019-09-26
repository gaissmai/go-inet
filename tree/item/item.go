// Package item defines the Itemer interface for tree items, maybe with payload and not just ip Blocks.
package item

// Itemer interface for tree items.
type Itemer interface {

	// Contains, defines the depth in the tree, parent child relationship.
	Contains(Itemer) bool

	// Compare, defines equality and sort order on same tree level, siblings relationship.
	Compare(Itemer) int
}
