// Package Tree is an implementation of a CIDR/Block prefix tree for fast IP lookup with longest-prefix-match.
// It is NOT a standard patricia-trie, this isn't possible for general blocks not represented by bitmasks.
package tree

import (
	"fmt"
	"io"
	"sort"

	"github.com/gaissmai/go-inet/inet"
)

type (
	// Tree, handle for the datastructure.
	Tree struct {
		// the entry point of the tree
		Root *Node
	}

	// Node in the tree, recursive data structure.
	Node struct {
		Item   *Item
		Parent *Node
		Childs []*Node
	}

	// Item, maybe with additonal payload, not just inet.Block.
	// It is intended that there is no Itemer interface.
	Item struct {
		// Block, methods Contains() and Compare() defines the position in the tree.
		Block inet.Block

		// payload for this tree item
		Payload interface{}

		// callback, helper func for generating the string
		StringCb func(Item) string
	}
)

// String calls back to the items sender, but only if StringCb is defined.
// The sender should better know how to format the payload, this library knows just the empty interface{}.
// If no callback for stringification is defined, return just the string for the inet.Block.
//
// It is intended that there is no Itemer interface.
func (item Item) String() string {
	if item.StringCb != nil {
		return item.StringCb(item)
	}
	// just return the String for inet.Block, don't know how to render the payload.
	return item.Block.String()
}

// New allocates a new tree and returns the pointer.
func New() *Tree {
	return &Tree{
		Root: &Node{
			Item:   nil, // tree has no payload in root-item slot
			Parent: nil, // parent of root-node is always nil
			Childs: nil, // here we start to insert items
		},
	}
}

// MustInsert is a helper that calls Insert and panics on error.
// It is intended for use in tree initializations.
func (t *Tree) MustInsert(items ...Item) {
	if err := t.Insert(items...); err != nil {
		panic(err)
	}
}

// Insert item(s) into the tree. Inserting a bulk of items is much faster
// than inserting unsorted single items in a loop.
//
// Returns error on duplicate items in the tree.
func (t *Tree) Insert(items ...Item) error {

	// sort before insert makes insertion much faster, no or less parent-child-relinking needed.
	sort.Slice(items, func(i, j int) bool { return items[i].Block.Compare(items[j].Block) < 0 })

	for i := range items {
		if err := t.Root.insert(items[i]); err != nil {
			return err
		}
	}

	return nil
}

// recursive work horse, use binary search on same level
// childs stay sorted after insert
func (node *Node) insert(input Item) error {

	// childs are sorted find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Block) >= 0 })

	// not at end of slice
	if idx < l {
		// don't insert dups
		if input.Block.Compare(node.Childs[idx].Item.Block) == 0 {
			return fmt.Errorf("duplicate item: %s", input)
		}
	}

	// not in front of slice, check if previous child contains new Item
	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(input.Block) {
			return child.insert(input)
		}
	}

	// add as new child on this level
	x := &Node{Item: &input, Parent: node, Childs: nil}

	// input is greater than all others and not contained, just append
	if idx == l {
		node.Childs = append(node.Childs, x)
		return nil
	}

	// buffer to build resorted childs
	buf := make([]*Node, 0, l+1)

	// copy [:idx] to buf
	buf = append(buf, node.Childs[:idx]...)

	// copy x to buf at [idx]
	buf = append(buf, x)

	// now handle [idx:]
	// resort if input contains next child...
	j := idx
	for {
		child := node.Childs[j]
		if input.Block.Contains(child.Item.Block) {
			// put old child under new Item
			if err := x.relinkNode(child); err != nil {
				return err
			}
			if j++; j < l {
				continue
			}
		}

		// childs are sorted, break after first child not being child of input
		break
	}

	// copy rest of childs to buf
	buf = append(buf, node.Childs[j:]...)

	node.Childs = buf
	return nil
}

// relinkNode with subtree/branch into tree
func (node *Node) relinkNode(input *Node) error {

	// childs are sorted find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Item.Block) >= 0 })

	// not at end of slice
	if idx < l {
		// don't insert dups
		if input.Item.Block.Compare(node.Childs[idx].Item.Block) == 0 {
			return fmt.Errorf("duplicate item: %s", input.Item)
		}
	}

	// not in front of slice, check if previous child contains this new node
	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(input.Item.Block) {
			return child.relinkNode(input)
		}
	}

	// add as new child on this level
	input.Parent = node

	// input is greater than all others and not contained, just append
	if idx == l {
		node.Childs = append(node.Childs, input)
		return nil
	}

	// buffer to build resorted childs
	buf := make([]*Node, 0, l+1)

	// copy [:idx] to buf
	buf = append(buf, node.Childs[:idx]...)

	// copy x to buf at [idx]
	buf = append(buf, input)

	// now handle [idx:]
	// resort if input contains next child...
	j := idx
	for {
		child := node.Childs[j]
		if input.Item.Block.Contains(child.Item.Block) {
			// recursive call, put next child in row under new input.Item
			if err := input.relinkNode(child); err != nil {
				panic(err)
			}
			if j++; j < l {
				continue
			}
		}

		// childs are sorted, break after first child not being child of input
		break
	}

	// copy rest of childs to buf
	buf = append(buf, node.Childs[j:]...)

	node.Childs = buf
	return nil
}

// RemoveBranch from tree. Returns true on success, false if not found.
func (t *Tree) RemoveBranch(item Item) bool {
	return t.Root.remove(item, true)
}

// Remove one item from tree, relink parent/child relation at the gap. Returns true on success,
// false if not found.
func (t *Tree) Remove(item Item) bool {
	return t.Root.remove(item, false)
}

// recursive work horse
func (node *Node) remove(input Item, andBranch bool) bool {

	// childs are sorted, find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Block) >= 0 })

	if idx != l && input.Block.Compare(node.Childs[idx].Item.Block) == 0 {
		// found child to remove at [idx]
		// delete this child [idx] from node
		child := node.Childs[idx]

		if idx < l-1 {
			node.Childs = append(node.Childs[:idx], node.Childs[idx+1:]...)
		} else {
			node.Childs = node.Childs[:idx]
		}

		// remove branch, stop here
		if andBranch {
			return true
		}

		// re-insert grandChilds from removed child into tree
		// just relinking parent-child not possible when blocks(ranges) overlaps.
		for _, grandChild := range child.Childs {

			// insert this grandchild at/under outer node
			if err := node.relinkNode(grandChild); err != nil {
				panic(err)
			}
		}

		return true
	}

	// pos in tree not found on this level
	// walk down if any child (before input, respect sort order) includes input
	for j := idx - 1; j >= 0; j-- {
		child := node.Childs[j]
		if child.Item.Block.Contains(input.Block) {
			return child.remove(input, andBranch)
		}
	}

	// not equal to any child and not contained in any child
	return false
}

// Contains reports whether the item is contained in any element of the tree.
// Just returns true or false and not the matching prefix,
// this is faster than a full Lookup for the longest-prefix-match.
func (t *Tree) Contains(item Item) bool {
	// just look in root childs, therefore much faster than a full tree lookup
	return t.Root.contains(item)
}

// returns true if item is contained in any node
func (node *Node) contains(query Item) bool {
	// find pos in childs on root level, binary search, childs are sorted
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(query.Block) >= 0 })

	if idx == 0 {
		return false
	}

	if idx < l {
		child := node.Childs[idx]
		// found by exact match?
		if child.Item.Block.Compare(query.Block) == 0 {
			return true
		}
	}

	// item before idx contains query?
	child := node.Childs[idx-1]
	return child.Item.Block.Contains(query.Block)
}

// Lookup item for longest prefix match in the tree.
// If not found, returns input argument and false.
func (t *Tree) Lookup(item Item) (Item, bool) {
	return t.Root.lookup(item)
}

// recursive work horse
func (node *Node) lookup(query Item) (Item, bool) {

	// find pos in childs on this level, binary search, childs are sorted
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(query.Block) >= 0 })

	if idx < l {
		child := node.Childs[idx]

		// found by exact match
		if child.Item.Block.Compare(query.Block) == 0 {
			return *child.Item, true
		}
	}

	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(query.Block) {
			return child.lookup(query)
		}
	}

	// no child path, no Item, we are at the root node
	if node.Item == nil {
		return query, false
	}

	// found by longest prefix match
	return *node.Item, true
}

// Fprint prints the ordered tree in ASCII graph.
//
// example:
//
//  ▼
//  ├─ ::/8.................   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 100::/8..............   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 200::/7..............   "Reserved by IETF     [RFC4048]"
//  ├─ 400::/6..............   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 800::/5..............   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 1000::/4.............   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 2000::/3.............   "Global Unicast       [RFC3513][RFC4291]"
//  │  ├─ 2000::/4.............  "Test"
//  │  └─ 3000::/4.............  "FREE"
//  ├─ 4000::/3.............   "Reserved by IETF     [RFC3513][RFC4291]"
//  ├─ 6000::/3.............   "Reserved by IETF     [RFC3513][RFC4291]"
func (t *Tree) Fprint(w io.Writer) {
	fmt.Fprintln(w, "▼")

	var walkAndPrint func(io.Writer, *Node, string)

	walkAndPrint = func(w io.Writer, node *Node, prefix string) {
		if node.Childs == nil {
			return
		}

		for i, child := range node.Childs {
			if i == len(node.Childs)-1 { // last child
				fmt.Fprintf(w, "%s%s\n", prefix+"└─ ", *child.Item)
				walkAndPrint(w, child, prefix+"   ")
			} else {
				fmt.Fprintf(w, "%s%s\n", prefix+"├─ ", *child.Item)
				walkAndPrint(w, child, prefix+"│  ")
			}
		}
	}

	walkAndPrint(w, t.Root, "")
}

// WalkFunc is the type of the function called for each node visited by Walk().
// The arguments to the WalkFunc are the current node in the tree
// and the depth, starting with 0.
//
// The Walk() stops if the WalkFunc returns an error.
type WalkFunc func(node *Node, depth int) error

// Walk the Tree starting at root in depth first order, calling walkFn for each node.
// At every node the walkFn is called with the node and the current depth as arguments.
// The walk stops if the walkFn returns an error not nil. The error is propagated by Walk() to the caller.
func (t *Tree) Walk(walkFn WalkFunc) error {

	// recursive work horse, declare ahead, recurse call below
	var walk func(*Node, WalkFunc, int) error

	walk = func(node *Node, walkFn WalkFunc, depth int) error {
		if node.Item != nil {
			if err := walkFn(node, depth); err != nil {
				return err
			}
		}

		// walk the childs
		for _, child := range node.Childs {
			if err := walk(child, walkFn, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	// just do it
	return walk(t.Root, walkFn, -1)
}
