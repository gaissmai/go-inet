// Package tree is an implementation of a CIDR/Block prefix tree for fast IP lookup with longest-prefix-match.
// It is NOT a standard patricia-trie, this isn't possible for general blocks not represented by bitmasks.
package tree

import (
	"fmt"
	"io"
	"sort"

	"github.com/gaissmai/go-inet/inet"
)

// Tree handle for the datastructure.
type Tree struct {
	// the entry point of the tree
	Root *Node
}

// Node in the tree, recursive data structure.
type Node struct {
	Item   *Item
	Parent *Node
	Childs []*Node
}

// Item in the node, maybe with additional payload, not just inet.Block.
// It is intended that there is no Itemer interface.
type Item struct {
	Block    inet.Block        // Block.Contains() and Block.Compare() define the position in the tree.
	Payload  interface{}       // payload for this tree item
	StringCb func(Item) string // callback, helper func for generating the string
}

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

	if len(items) > 1 {
		// sort before insert makes insertion much faster, no or less parent-child-relinking needed.
		sort.Slice(items, func(i, j int) bool { return items[i].Block.Compare(items[j].Block) < 0 })
	}

	for i := range items {
		if err := t.Root.insertItem(items[i]); err != nil {
			return err
		}
	}

	return nil
}

// RemoveBranch from tree. Returns error on failure or not found.
func (t *Tree) RemoveBranch(item Item) error {
	return t.Root.remove(item, true)
}

// Remove one item from tree, relink parent/child relation at the gap. Returns error if not found.
func (t *Tree) Remove(item Item) error {
	return t.Root.remove(item, false)
}

// Contains reports whether the item is contained in any element of the tree.
// Just returns true or false and not the matching prefix,
// this is faster than a full Lookup for the longest-prefix-match.
func (t *Tree) Contains(item Item) bool {
	// just look in root childs, therefore maybe faster than a full tree lookup
	return t.Root.contains(item)
}

// Lookup item for longest prefix match in the tree.
// If not found, returns input argument and false.
func (t *Tree) Lookup(item Item) (Item, bool) {
	return t.Root.lookup(item)
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
		if len(node.Childs) == 0 {
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

	// recursive func, declare ahead, recurse call below
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

// Len returns the number of items in the tree.
func (t *Tree) Len() int {
	var nodes int
	_ = t.Walk(func(*Node, int) error { nodes++; return nil })
	return nodes
}

// insert single item, make a node and link node in tree
func (node *Node) insertItem(item Item) error {

	// convert item to node and use general insertNode
	newNode := &Node{Item: &item, Parent: nil, Childs: nil}

	return node.insertNode(newNode)
}

// insertNode, one method for new nodes and parent-child relinking
// recursive descent
func (node *Node) insertNode(input *Node) error {

	// childs are sorted find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Item.Block) >= 0 })

	// not at end of slice
	if idx < l && input.Item.Block.Compare(node.Childs[idx].Item.Block) == 0 {
		// don't insert dups
		return fmt.Errorf("duplicate item: %s", input.Item)
	}

	// not in front of slice, check if previous child contains this new node
	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(input.Item.Block) {
			// recursive descent
			return child.insertNode(input)
		}
	}

	// add as new child on this level
	input.Parent = node

	// input is greater than all others and not contained in child before, just append
	if idx == l {
		node.Childs = append(node.Childs, input)
		return nil
	}

	// input child somewhere in the 'middle'

	// buffer to build reordered childs
	// buf = ([:idx], input, [idx:])
	buf := make([]*Node, 0, l+1)

	// copy til [:idx] to new childs buf
	buf = append(buf, node.Childs[:idx]...)

	// append new input node
	buf = append(buf, input)

	// now handle the rest of childs [idx:]
	// relink if new input node contains next child(s)... in row
	j := idx
	for ; j < l; j++ {
		child := node.Childs[j]

		if input.Item.Block.Contains(child.Item.Block) {
			// recursive descent, relink next child in row
			if err := input.insertNode(child); err != nil {
				return err
			}
			continue
		}
		// childs are sorted, break after first child not being child of input
		break
	}

	// now copy rest of childs to buf
	buf = append(buf, node.Childs[j:]...)
	node.Childs = buf

	return nil
}

// recursive descent
func (node *Node) remove(input Item, delBranch bool) error {

	// childs are sorted, find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Block) >= 0 })

	// check for exact match
	if idx != l && input.Block.Compare(node.Childs[idx].Item.Block) == 0 {

		// save for relinking of grandchilds, delete this child at idx from node
		match := node.Childs[idx]

		// ####
		// cut elem without memory leak, see golang wiki slice tricks
		//
		// idx is not at end of slice, somewhere in the 'middle'
		if idx < l-1 {
			copy(node.Childs[idx:], node.Childs[idx+1:])
		}
		// reset last elem in slice to nil, can be garbage collected, make no memory leak
		node.Childs[l-1] = nil

		// cut this last entry from slice
		node.Childs = node.Childs[:l-1]

		// if we are in remove branch mode, stop here, no relinking of grand childs
		if delBranch {
			return nil
		}

		// re-insert grandChilds from deleted child into tree
		for _, grandChild := range match.Childs {

			// insert this grandchild, start at node
			if err := node.insertNode(grandChild); err != nil {
				return err
			}
		}

		return nil
	}

	// no exact match at this level, check if child before idx contains the input?
	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(input.Block) {
			return child.remove(input, delBranch)
		}
	}

	// not equal to any child and not contained in any child
	return fmt.Errorf("remove, item not found: %s", input)
}

// shallow test in root childs
func (node *Node) contains(query Item) bool {
	l := len(node.Childs)

	// find pos in root-childs, binary search, childs are always sorted
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(query.Block) >= 0 })

	// search index may be at l, take care for index panics
	if idx < l {

		// child at idx may be equal to item
		if node.Childs[idx].Item.Block.Compare(query.Block) == 0 {
			return true
		}
	}

	// search index may be 0, take care for index panics
	if idx > 0 {
		// child before idx may contain the item
		if node.Childs[idx-1].Item.Block.Contains(query.Block) {
			return true
		}
	}

	return false
}

// find exact match or longest-prefix-match
// recursive descent algo
func (node *Node) lookup(query Item) (Item, bool) {

	l := len(node.Childs)

	// find pos in childs on this level, binary search, childs are sorted
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(query.Block) >= 0 })

	// found by exact match?
	if idx < l {
		if node.Childs[idx].Item.Block.Compare(query.Block) == 0 {
			return *node.Childs[idx].Item, true
		}
	}

	// search index may be 0, take care for index panics
	if idx > 0 {
		// make alias, better to read or debug
		child := node.Childs[idx-1]

		if child.Item.Block.Contains(query.Block) {
			// recursive descent
			return child.lookup(query)
		}
	}

	// no child contains item and node has no Item, we are at the root node
	if node.Item == nil {
		return query, false
	}

	// ok, this node is the end of recurse descent, return longest prefix match
	return *node.Item, true
}
