package inet

import (
	"fmt"
	"io"
	"sort"
)

// NewTree allocates a new tree and returns the pointer.
func NewTree() *Tree {
	return &Tree{
		Root: &Node{
			Item:   nil, // multi-root tree has no payload in root-item slot
			Parent: nil, // parent of root-node is always nil
			Childs: nil,
		},
	}
}

// Insert one item into the tree. The position within the tree is defined
// by the Contains() method, part of the Itemer interface .
func (t *Tree) Insert(b Itemer) *Tree {
	// parent of root is nil
	t.Root.insert(nil, b)
	return t
}

// recursive work horse
func (n *Node) insert(p *Node, b Itemer) {

	// found pos, item is nil, insert payload, but not at root level (t.Root.Parent == nil)
	if n.Item == nil && p != nil {
		n.Item = &b
		n.Parent = p
		return
	}

	// find pos, walk down
	for _, c := range n.Childs {
		// check for dups
		if b.Compare(*c.Item) == 0 {
			// dup found, don't insert
			return
		}

		// go down
		if (*c.Item).Contains(b) {
			c.insert(n, b)
			return
		}
	}

	// add as new child on this level
	newNode := &Node{Item: &b, Parent: n, Childs: nil}

	// if any child in this level is subset of new item, rearrange
	keep := make([]*Node, 0, len(n.Childs))

	for _, c := range n.Childs {
		if b.Contains(*c.Item) {
			// put child under new Item
			c.Parent = newNode
			newNode.Childs = append(newNode.Childs, c)
		} else {
			// keep child
			keep = append(keep, c)
		}
	}

	n.Childs = append(keep, newNode)
}

// Remove one item from tree, relink parent/child relation at the gap. Returns true on success,
// false if not found.
func (t *Tree) Remove(b Itemer) bool {
	return t.Root.remove(b)
}

// recursive work horse
func (n *Node) remove(b Itemer) bool {
	// found pos
	if n.Item != nil && (*n.Item).Compare(b) == 0 {

		// remove this node from parent childs, keep siblings
		keep := make([]*Node, 0, len(n.Parent.Childs))
		for _, c := range n.Parent.Childs {

			// not me, just a sibling
			if (*c.Item).Compare(b) != 0 {
				keep = append(keep, c)
			}

		}
		n.Parent.Childs = keep

		// relink the childs to parent
		for _, c := range n.Childs {
			c.Parent = n.Parent
			n.Parent.Childs = append(n.Parent.Childs, c)
		}

		return true
	}

	// not found, walk down
	for _, c := range n.Childs {
		if (*c.Item).Contains(b) || (*c.Item).Compare(b) == 0 {
			return c.remove(b)
		}
	}
	return false
}

// Lookup item for longest prefix match in the tree.
// If not found, returns input argument and false.
func (t *Tree) Lookup(b Itemer) (Itemer, bool) {
	return t.Root.lookup(b)
}

// recursive work horse
func (n *Node) lookup(b Itemer) (Itemer, bool) {

	// found by equality
	if n.Item != nil && (*n.Item).Compare(b) == 0 {
		return b, true
	}

	// not found, walk down
	for _, c := range n.Childs {
		if (*c.Item).Contains(b) || (*c.Item).Compare(b) == 0 {
			return c.lookup(b)
		}
	}

	// no child path, no item and no parent, we are at the root
	if n.Parent == nil || n.Item == nil {
		return b, false
	}

	// found by longest prefix match
	return *n.Item, true
}

// Fprint prints the ordered tree in ASCII graph to io.Writer.
// The items should implement the Stringer interface for readable output.
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

	walkAndPrint = func(w io.Writer, n *Node, prefix string) {
		if n.Childs == nil {
			return
		}

		// sort the childs in ascending order before printing
		// use Compare from Itemer interface
		sort.Slice(n.Childs, func(i, j int) bool {
			return (*n.Childs[i].Item).Compare(*n.Childs[j].Item) < 0
		})

		for i, c := range n.Childs {
			if i == len(n.Childs)-1 { // last child
				fmt.Fprintf(w, "%s%s\n", prefix+"└─ ", *c.Item)
				walkAndPrint(w, c, prefix+"   ")
			} else {
				fmt.Fprintf(w, "%s%s\n", prefix+"├─ ", *c.Item)
				walkAndPrint(w, c, prefix+"│  ")
			}
		}
	}

	walkAndPrint(w, t.Root, "")
}

// WalkFunc is the type of the function called for each node visited by Walk().
// The arguments to the WalkFunc are the current node in the tree
// and the depth, starting by 0.
//
// The Walk() stops if the WalkFunc returns an error.
type WalkFunc func(n *Node, depth int) error

// Walk the Tree starting at root, calling walkFn for each node.
// The walk is in sorted order if requested.
// At every node the walkFn is called with the node and the current depth as arguments.
// The walk stops if the walkFn returns an error. The error is propagated by Walk() to the caller.
func (t *Tree) Walk(walkFn WalkFunc, sorted bool) error {

	// recursive work horse, declare ahead, recurse call below
	var walk func(*Node, WalkFunc, int) error

	walk = func(n *Node, walkFn WalkFunc, depth int) error {
		if n.Item != nil {
			if err := walkFn(n, depth); err != nil {
				return err
			}
		}

		// sort the childs in ascending order before walking
		if sorted {
			lessFn := func(i, j int) bool {
				return (*n.Childs[i].Item).Compare(*n.Childs[j].Item) < 0
			}
			if !sort.SliceIsSorted(n.Childs, lessFn) {
				sort.Slice(n.Childs, lessFn)
			}
		}

		// now walk the childs
		for _, c := range n.Childs {
			if err := walk(c, walkFn, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	// just do it
	return walk(t.Root, walkFn, -1)
}
