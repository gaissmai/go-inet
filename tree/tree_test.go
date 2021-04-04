package tree

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
)

// simple test interval
type ival struct {
	lo, hi int
}

// implementing tree.Interface

// Equal
func (a ival) Equals(i Interface) bool {
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
	tree, _ := New(nil)

	if s := tree.String(); s != "" {
		t.Errorf("tree.String() = %v, want \"\"", s)
	}

	if l := tree.Len(); l != 0 {
		t.Errorf("tree.Len() = %v, want 0", l)
	}

	if i := tree.Lookup(nil); i != nil {
		t.Errorf("tree.Lookup(nil) = %v, want nil", i)
	}

	for _, tt := range []struct {
		f    func(Interface) Interface
		item Interface
		name string
	}{
		{tree.Superset, nil, "tree.Superset"},
		{tree.Lookup, nil, "tree.Lookup"},
		//
		{tree.Superset, ival{}, "tree.Superset"},
		{tree.Lookup, ival{}, "tree.Lookup"},
		//
		{tree.Superset, ival{1, 2}, "tree.Superset"},
		{tree.Lookup, ival{1, 2}, "tree.Lookup"},
	} {
		if m := tt.f(tt.item); m != nil {
			t.Errorf("%s(%s), got m = %v, expected <nil>", tt.name, tt.item, m)
		}
	}

	if i := tree.Walk(nil); i != nil {
		t.Errorf("tree.Walk(nil) = %v, want nil", i)
	}

}

func TestTreeNewWithDups(t *testing.T) {

	tree, err := New([]Interface{ival{0, 0}, ival{6, 10}, ival{42, 4242}, ival{42, 4242}})
	if err == nil {
		t.Errorf("expected error, got: %v", err)
	}
	if d := tree.Duplicates(); d == nil {
		t.Errorf("expected dups, got: %v", d)
	}
	if d := tree.Duplicates()[0]; !d.Equals(Interface(ival{42, 4242})) {
		t.Errorf("expected: %v, got: %v", ival{42, 4242}, d)
	}

}

func TestTreeLookup(t *testing.T) {

	is := []Interface{
		ival{1, 100},
		ival{45, 60},
	}

	tree, err := New(is)
	if err != nil {
		t.Error(err)
	}

	item := ival{0, 6}
	if got := tree.Lookup(item); got != nil {
		t.Errorf("Lookup(%v) = %v, want %v", item, got, nil)
	}

	item = ival{47, 62}
	if got := tree.Lookup(item); !got.Equals(ival{1, 100}) {
		t.Errorf("Lookup(%v) = %v, want %v", item, got, ival{1, 100})
	}
}

func TestTreeWalk(t *testing.T) {
	var tree *Tree
	if i := tree.Walk(nil); i != nil {
		t.Errorf("tree.Walk(nil) = %v, want nil", i)
	}

	is := generateIvals(21)
	tree, _ = New(is)

	var count int
	var maxDepth int
	var maxChilds int
	var itemWithMaxChilds Interface

	// as closure
	walkFn := func(d int, item, parent Interface, childs []Interface) error {
		count++
		if d > maxDepth {
			maxDepth = d
		}
		if len(childs) > maxChilds {
			maxChilds = len(childs)
			itemWithMaxChilds = item
		}
		return nil
	}
	if err := tree.Walk(walkFn); err != nil {
		t.Errorf("Walk returns error: %v", err)
	}

	if count != 21 {
		t.Errorf("Walk, count items, expected 21, got: %v", count)
	}

	if maxDepth != 6 {
		t.Errorf("Walk, maxDepth, expected 6, got: %v", maxDepth)
	}

	if maxChilds != 3 {
		t.Errorf("Walk, maxChilds, expected 3, got: %v", maxChilds)
	}

	exp := ival{0, 20}
	if !exp.Equals(itemWithMaxChilds) {
		t.Errorf("Walk, itemWithMaxChilds, expected 0...20, got: %v", itemWithMaxChilds)
	}

	// now with error
	walkFn2 := func(d int, item, parent Interface, childs []Interface) error {
		if d > 4 {
			return errors.New("to deep")
		}
		return nil
	}

	if err := tree.Walk(walkFn2); err == nil {
		t.Errorf("Walk, expected error, got: %v", err)
	}
}

func TestTreeRandom(t *testing.T) {
	is := generateIvals(10_000)
	tree, _ := New(is)

	if len(is) != tree.Len() {
		t.Errorf("Len(), got %v, expected %v", tree.Len(), len(is))
	}

	rand.Shuffle(len(is), func(i, j int) { is[i], is[j] = is[j], is[i] })

	for _, item := range is {
		if m := tree.Lookup(item); m == nil {
			t.Errorf("Lookup(%v), got %v", item, m)
		}
		if m := tree.Superset(item); m == nil {
			t.Errorf("Superset(%v), got %v", item, m)
		}
	}
}
