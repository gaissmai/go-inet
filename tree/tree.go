package tree

import (
	"errors"
	"sort"
	"strings"
)

// parent of all childs
const root = -1

// An Interface for various methods on intervals.
type Interface interface {
	// Covers returns true if and only if the receiver truly covers item.
	// When receiver covers item, they cannot be equal, thus Equal then will return false!
	// If Covers is true, Less also must be true.
	Covers(Interface) bool

	// Less reports whether the receiver should be sorted before the item.
	// HINT: All supersets must be ordered to the left of their subsets!
	Less(Interface) bool

	// Equals reports whether receiver and item are equal.
	Equals(Interface) bool

	// Stringer interface
	String() string
}

// Tree partially implements an interval tree.
type Tree struct {
	// the sorted items, immutable, stored as slice, not as tree
	items []Interface

	// top-down parentIdx -> []childIdx tree
	tree map[int][]int

	// the duplicate items
	dups []Interface
}

// Duplicates returns the conflicting items. Returns nil if there was no error during New().
func (t *Tree) Duplicates() []Interface {
	return t.dups
}

// Len returns the number of items in tree.
func (t *Tree) Len() int {
	return len(t.items)
}

// New builds and returns an immutable tree.
// Returns an error != nil on duplicate items.
func New(items []Interface) (*Tree, error) {
	t := &Tree{}
	if items == nil {
		return t, nil
	}

	t.items = make([]Interface, len(items))
	t.tree = make(map[int][]int)

	// copy/clone and sort input, decouple from caller
	copy(t.items, items)
	sort.Slice(t.items, func(i, j int) bool { return t.items[i].Less(t.items[j]) })

	// items are sorted, build the index tree, O(n), bail out on duplicates
	for i := range t.items {

		// collect the dups
		if i > 0 && t.items[i-1].Equals(t.items[i]) {
			t.dups = append(t.dups, t.items[i])
			continue
		}
		t.buildIndexTree(root, i)
	}

	if t.dups != nil {
		return t, errors.New("some items are duplicate")
	}

	return t, nil
}

// buildIndexTree, parent->child map, rec-descent algo.
// Just building the tree with the slice indices, the items itself are not moved.
func (t *Tree) buildIndexTree(p, c int) {

	// if child index slice is empty, just append the childs index
	if t.tree[p] == nil {
		t.tree[p] = append(t.tree[p], c)
		return
	}

	// everything is sorted, just compare with last child index
	cs := t.tree[p] // dereference

	// get last child index of this parent
	cLast := cs[len(cs)-1]

	// item is covered by last child, rec-descent down in tree
	// last child is new parent
	if t.items[cLast].Covers(t.items[c]) {
		t.buildIndexTree(cLast, c)
		return
	}

	// not covered by any child, just append at this level the child index
	t.tree[p] = append(t.tree[p], c)
}

// Lookup returns the item itself or the *smallest* superset (bottom-up).
// If item is not covered at all by tree, then the returned item is nil.
//
// Example: Can be used in IP-ranges or IP-CIDRs to find the so called longest-prefix-match.
func (t *Tree) Lookup(item Interface) Interface {
	if t.items == nil || item == nil {
		return nil
	}
	// rec-descent
	return t.lookup(root, item)
}

func (t *Tree) lookup(p int, item Interface) Interface {
	// dereference
	cs := t.tree[p]

	// find pos in slice on this level
	i := sort.Search(len(cs), func(i int) bool { return item.Less(t.items[cs[i]]) })

	// child before idx may be equal or covers item
	if i > 0 {
		i--
		if t.items[cs[i]].Equals(item) {
			return t.items[cs[i]]
		}
		if t.items[cs[i]].Covers(item) {
			return t.lookup(cs[i], item)
		}
	}

	// return parent at this level
	if p != root {
		return t.items[p]
	}
	return nil
}

// Superset returns the *biggest* superset (top-down) or the item itself.
// Find first interval covering item in root level.
// If item is not contained at all in tree, then the returned item is nil.
// Extremely degraded trees with heavy overlaps result in O(n).
func (t *Tree) Superset(item Interface) Interface {
	if item == nil {
		return nil
	}

	// find first item with O(n) in root level
	for _, v := range t.tree[root] {
		if t.items[v].Equals(item) || t.items[v].Covers(item) {
			return t.items[v]
		}
	}

	return nil
}

// String returns the ordered tree as a directory graph.
// The items are stringified using their fmt.Stringer interface.
func (t *Tree) String() string {
	str := t.walkAndStringify(root, new(strings.Builder), "").String()

	if str == "" {
		return ""
	}
	return "▼\n" + str
}

// walkAndStringify rec-descent, top-down
func (t *Tree) walkAndStringify(p int, buf *strings.Builder, pad string) *strings.Builder {
	cs := t.tree[p]
	l := len(cs)

	// stop condition, no more childs
	if l == 0 {
		return buf
	}

	// !!! stop before last child (<= l-2)
	i := 0
	for ; i <= l-2; i++ {
		idx := cs[i] // dereference

		buf.WriteString(pad + "├─ " + t.items[idx].String() + "\n")
		buf = t.walkAndStringify(idx, buf, pad+"│  ")
	}

	// treat last child special
	idx := cs[i] // dereference

	buf.WriteString(pad + "└─ " + t.items[idx].String() + "\n")
	return t.walkAndStringify(idx, buf, pad+"   ")
}

// WalkTreeFunc is the type of the function called by WalkTree to visit each item.
//
// depth is 0 for root items.
//
// item is the visited item.
//
// parent is nil for root items.
//
// childs is the slice of direct descendants, childs is nil for leaf items.
//
// If the function returns a non-nil error, Walk stops and returns that error.
type WalkTreeFunc func(depth int, item, parent Interface, childs []Interface) error

// String returns the ordered tree as a directory graph.
// The items are stringified using their fmt.Stringer interface.
func (t *Tree) Walk(fn WalkTreeFunc) error {
	if t == nil {
		return nil
	}

	// for all child_idxs of the root item...
	for _, idx := range t.tree[root] {
		if err := t.walk(fn, 0, idx, root); err != nil {
			return err
		}
	}
	return nil
}

func (t *Tree) walk(fn WalkTreeFunc, d, idx, p int) error {
	var item, parent Interface
	var childs []Interface

	item = t.items[idx]

	if p != root {
		parent = t.items[p]
	}

	c_idxs := t.tree[idx]
	for _, c := range c_idxs {
		childs = append(childs, t.items[c])
	}

	// visitor callback
	if err := fn(d, item, parent, childs); err != nil {
		return err
	}

	// rec-descent
	for _, c := range c_idxs {
		if err := t.walk(fn, d+1, c, idx); err != nil {
			return err
		}
	}

	return nil
}
