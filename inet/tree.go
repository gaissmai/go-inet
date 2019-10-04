package inet

import (
	"fmt"
	"io"
	"reflect"
	"sort"
)

// NewTree allocates a new tree and returns the pointer.
func NewTree() *Tree {
	return &Tree{
		Root: &Node{
			Item:   nil, // multi-root tree has no payload in root-item slot
			Parent: nil, // parent of root-node is always nil
			Childs: nil, // here we start to insert items
		},
	}
}

// InsertBulk takes a slice of []items, sorts the values and inserts it into the tree.
// If the input arg is no slice of items implementing the Itemer interface,
// the method will panic.
//
// It is a convenience method. The user can also sort the slice itself and
// insert the elements in a loop with Insert()
func (t *Tree) InsertBulk(i interface{}) *Tree {
	slice := reflect.ValueOf(i)
	if slice.Kind() != reflect.Slice {
		panic("input isn't a slice of items")
	}
	l := slice.Len()
	items := make([]Itemer, l)
	for i := 0; i < l; i++ {
		item, ok := slice.Index(i).Interface().(Itemer)
		if !ok {
			panic("input items don't implement the Itemer interface")
		}
		items[i] = item
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Compare(items[j]) < 0 })
	for i := range items {
		t.Root.insert(items[i])
	}
	return t
}

// Insert one item into the tree. The position within the tree is defined
// by the Contains() and Compare() methods, part of the Itemer interface .
//
// If you insert many values you should sort them first.
func (t *Tree) Insert(b Itemer) *Tree {
	// parent of root is nil
	t.Root.insert(b)
	return t
}

// recursive work horse, use binary search on same level
// childs stay sorted after insert
func (n *Node) insert(b Itemer) {

	// childs are sorted
	// find pos in childs on this level, binary search
	i := sort.Search(len(n.Childs), func(i int) bool { return (*n.Childs[i].Item).Compare(b) >= 0 })

	l := len(n.Childs)
	// not at end of slice
	if i < l {
		// don't insert dups
		if b.Compare(*n.Childs[i].Item) == 0 {
			return
		}
	}

	// not in front of slice, check if previous child contains new Item
	if i > 0 {
		c := n.Childs[i-1]
		if (*c.Item).Contains(b) {
			c.insert(b)
			return
		}
	}

	// add as new child on this level
	x := &Node{Item: &b, Parent: n, Childs: nil}

	// b is greater than all others and not contained, just append
	if i == l {
		n.Childs = append(n.Childs, x)
		return
	}

	// buffer to build resorted childs
	buf := make([]*Node, 0, l+1)

	// copy [:i] to buf
	buf = append(buf, n.Childs[:i]...)

	// copy x to buf at [i]
	buf = append(buf, x)

	// now handle [i:]
	// resort if b contains next child...
	j := i
	for {
		c := n.Childs[j]
		if b.Contains(*c.Item) {
			// put old child under new Item
			x.Childs = append(x.Childs, c)
			c.Parent = x
			if j++; j < l {
				continue
			}
		}

		// childs are sorted, break after first child not being child of b
		break
	}

	// copy rest of childs to buf
	buf = append(buf, n.Childs[j:]...)

	n.Childs = buf
}

// Remove one item from tree, relink parent/child relation at the gap. Returns true on success,
// false if not found.
func (t *Tree) Remove(b Itemer) bool {
	return t.Root.remove(b)
}

// recursive work horse
func (n *Node) remove(b Itemer) bool {

	// childs are sorted
	// find pos in childs on this level, binary search
	i := sort.Search(len(n.Childs), func(i int) bool { return (*n.Childs[i].Item).Compare(b) >= 0 })

	l := len(n.Childs)

	if i != l && b.Compare(*n.Childs[i].Item) == 0 {
		// found child to remove at [i]
		// delete this child [i] from node
		c := n.Childs[i]

		if i < l-1 {
			n.Childs = append(n.Childs[:i], n.Childs[i+1:]...)
		} else {
			n.Childs = n.Childs[:i]
		}

		// re-insert grandchilds into tree at this node
		// just relinking of parent-child links not always possible
		// there may be some overlaps with containment edge cases
		// reinserting is safe
		var walk func(*Node)
		walk = func(c *Node) {
			for _, gc := range c.Childs {
				n.insert(*gc.Item)
				walk(gc)
			}
		}

		walk(c)
		return true
	}

	// pos in tree not found on this level
	// walk down if any child (before b, respect sort order) includes b
	for j := i - 1; j >= 0; j-- {
		c := n.Childs[j]
		if (*c.Item).Contains(b) {
			return c.remove(b)
		}
	}

	// not equal to any child and not contained in any child
	return false
}

// Lookup item for longest prefix match in the tree.
// If not found, returns input argument and false.
func (t *Tree) Lookup(b Itemer) (Itemer, bool) {
	return t.Root.lookup(b)
}

// recursive work horse
func (n *Node) lookup(b Itemer) (Itemer, bool) {

	// find pos in childs on this level, binary search
	// childs are sorted
	i := sort.Search(len(n.Childs), func(i int) bool { return (*n.Childs[i].Item).Compare(b) >= 0 })
	l := len(n.Childs)

	if i < l {
		c := n.Childs[i]

		// found by exact match
		if (*c.Item).Compare(b) == 0 {
			return *c.Item, true
		}
	}

	if i > 0 {
		c := n.Childs[i-1]
		if (*c.Item).Contains(b) {
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
// At every node the walkFn is called with the node and the current depth as arguments.
// The walk stops if the walkFn returns an error not nil. The error is propagated by Walk() to the caller.
func (t *Tree) Walk(walkFn WalkFunc) error {

	// recursive work horse, declare ahead, recurse call below
	var walk func(*Node, WalkFunc, int) error

	walk = func(n *Node, walkFn WalkFunc, depth int) error {
		if n.Item != nil {
			if err := walkFn(n, depth); err != nil {
				return err
			}
		}

		// walk the childs
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
