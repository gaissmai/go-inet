package tree

import (
	"fmt"
	"math/rand"
	"testing"
)

type ival struct {
	lo, hi int
}

// Equal
func (a ival) Equal(i Interface) bool {
	b := i.(ival)
	return a == b
}

// Covers
func (a ival) Covers(i Interface) bool {
	b := i.(ival)
	if a == b {
		return false
	}
	return a.lo <= b.lo && a.hi >= b.hi
}

// Less
func (a ival) Less(i Interface) bool {
	b := i.(ival)
	if a == b {
		return false
	}
	if a.lo == b.lo {
		// HINT: sort containers to the left!
		return a.hi > b.hi
	}
	return a.lo < b.lo
}

// fmt.Stringer
func (a ival) String() string {
	return fmt.Sprintf("%d...%d", a.lo, a.hi)
}

func generateIvals(n int) []Interface {
	set := make(map[ival]int, n)
	for i := 0; i < n; i++ {
		a := rand.Intn(n)
		b := rand.Intn(n)
		if a > b {
			a, b = b, a
		}
		iv := ival{a, b}
		set[iv]++
	}
	is := make([]Interface, 0, len(set))
	for k := range set {
		is = append(is, k)
	}
	return is
}

func TestTreeNil(t *testing.T) {
	tree, _ := NewTree(nil)

	if s := tree.String(); s != "" {
		t.Errorf("tree.String() = %v, want \"\"", s)
	}

	if l := tree.Len(); l != 0 {
		t.Errorf("tree.Len() = %v, want 0", l)
	}

	// func(Interface) (Interface, bool)
	for _, tt := range []struct {
		f    func(Interface) Interface
		item Interface
		name string
	}{
		{tree.Superset, nil, "tree.Contains"},
		{tree.Lookup, nil, "tree.Lookup"},
		//
		{tree.Superset, ival{}, "tree.Contains"},
		{tree.Lookup, ival{}, "tree.Lookup"},
		//
		{tree.Superset, ival{1, 2}, "tree.Contains"},
		{tree.Lookup, ival{1, 2}, "tree.Lookup"},
	} {
		if m := tt.f(tt.item); m != nil {
			t.Errorf("%s(%s), got m = %v, expected <nil>", tt.name, tt.item, m)
		}
	}

}

func TestTreeInsertDup(t *testing.T) {

	_, err := NewTree([]Interface{ival{42, 4242}, ival{42, 4242}})
	if err == nil {
		t.Errorf("expected dups, got nil")
	}
}

func TestTreeRandom(t *testing.T) {
	is := generateIvals(10_000)
	tree, _ := NewTree(is)

	if len(is) != tree.Len() {
		t.Errorf("Len(), got %v, expected %v", tree.Len(), len(is))
	}

	rand.Shuffle(len(is), func(i, j int) { is[i], is[j] = is[j], is[i] })

	for _, item := range is {
		if m := tree.Lookup(item); m == nil {
			t.Errorf("Lookup(%v), got %v", item, m)
		}
		if m := tree.Superset(item); m == nil {
			t.Errorf("Contains(%v), got %v", item, m)
		}
	}
}

// ▼
// ├─ 0...15
// │  └─ 0...6
// ├─ 1...245
// │  ├─ 1...87
// │  │  └─ 1...18
// │  ├─ 2...89
// │  └─ 4...211
// │     └─ 5...140
// │        └─ 5...66
// └─ 7...247
//    └─ 7...206
//       └─ 7...88
//          ├─ 7...59
//          └─ 8...74
//             └─ 8...58

var sVals = []Interface{
	ival{0, 15}, ival{0, 6}, ival{1, 245}, ival{1, 87}, ival{1, 18}, ival{2, 89}, ival{4, 211}, ival{5, 140},
	ival{5, 66}, ival{7, 247}, ival{7, 206}, ival{7, 88}, ival{7, 59}, ival{8, 74}, ival{8, 58},
}
