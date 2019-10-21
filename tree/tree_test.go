package tree

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/internal"
)

func TestTreeInsertBulk(t *testing.T) {
	n := 20000
	cidrs := internal.GenBlockMixed(n)
	ranges := internal.GenRangeMixed(n)

	blocks := make([]inet.Block, 0, len(cidrs)+len(ranges))
	blocks = append(blocks, cidrs...)
	blocks = append(blocks, ranges...)

	items := make([]Item, len(blocks))
	for i := range blocks {
		items[i] = Item{Block: blocks[i]}
	}

	tr := New()
	err := tr.Insert(items...)
	if err != nil {
		t.Errorf("Insert error: %s", err)
	}

	got := tr.Len()
	if got != 2*n {
		t.Errorf("Len() = %d, want %d", got, n)
	}
}

func TestTreeInsertBulkRemoveRandom(t *testing.T) {
	n := 20000
	cidrs := internal.GenBlockMixed(n)
	ranges := internal.GenRangeMixed(n)

	blocks := make([]inet.Block, 0, len(cidrs)+len(ranges))
	blocks = append(blocks, cidrs...)
	blocks = append(blocks, ranges...)

	items := make([]Item, len(blocks))
	for i := range blocks {
		items[i] = Item{Block: blocks[i]}
	}

	tr := New()
	err := tr.Insert(items...)
	if err != nil {
		t.Errorf("Insert error: %s", err)
	}

	r := rand.New(rand.NewSource(42))
	set := map[int]bool{}
	for m := 0; m < n/100; {
		i := r.Intn(n)
		if set[i] {
			continue
		}
		m++
		set[i] = true
		if err := tr.Remove(items[i]); err != nil {
			t.Errorf("Remove error: %s", err)
		}
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeInsertDups(t *testing.T) {
	n := 200
	cidrs := internal.GenBlockMixed(n)
	ranges := internal.GenRangeMixed(n)

	blocks := make([]inet.Block, 0, len(cidrs)+len(ranges))
	blocks = append(blocks, cidrs...)
	blocks = append(blocks, ranges...)
	// make dups
	blocks = append(blocks, blocks...)

	items := make([]Item, len(blocks))
	for i := range blocks {
		items[i] = Item{Block: blocks[i]}
	}

	tr := New()
	err := tr.Insert(items...)
	if err != nil && !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("Insert error: %s", err)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeInsertNothing(t *testing.T) {
	tr := New()
	tr.MustInsert()

	got := new(strings.Builder)
	tr.Fprint(got)

	want := `▼
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeLookupMissing(t *testing.T) {
	s1 := Item{Block: inet.MustBlock("0.0.0.0/0")}
	s2 := Item{Block: inet.MustBlock("2001:db8::/32")}

	tr := New()
	tr.MustInsert(s1)

	_, ok := tr.Lookup(s2)
	if ok {
		t.Errorf("Lookup(%s) got %t, want %t", s2, ok, false)
	}
}

func TestTreeLookup(t *testing.T) {
	s1 := Item{Block: inet.MustBlock("0.0.0.0/0")}

	tr := New()
	tr.MustInsert(s1)

	got, ok := tr.Lookup(s1)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", s1, ok, true)
	}

	if got.Block.Compare(s1.Block) != 0 {
		t.Errorf("tr.Lookup(%v), got: %v, want: %v", s1, got, s1)
	}
}

func TestTreeLookupLPM(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	look := Item{inet.MustBlock("0.0.0.0/32"), nil, nil}
	want := inet.MustBlock("0.0.0.0/10")

	got, ok := tr.Lookup(look)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", look, ok, true)
	}

	if got.Block.Compare(want) != 0 {
		t.Errorf("tr.Lookup(%v), got: %v, want: %v", look, got, want)
	}
}

func TestTreeWalk(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/127",
		"2001:7c0:900:1c2::1/128",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	got := new(strings.Builder)

	var maxDepth int
	var maxWidth int

	var walkFn WalkFunc = func(n *Node, depth int) error {
		if depth > maxDepth {
			maxDepth = depth
		}
		if l := len(n.Childs); l > maxWidth {
			maxWidth = l
		}

		pfx := strings.Repeat("|", depth)
		fmt.Fprintf(got, "%s-->%s\n", pfx, *n.Item)
		return nil
	}

	err := tr.Walk(walkFn)

	if err != nil {
		t.Errorf("%v\n", err)
	}

	want :=
		`-->0.0.0.0/0
|-->0.0.0.0/8
||-->0.0.0.0/10
|-->1.0.0.0/8
|-->5.0.0.0/8
|-->10.0.0.0-10.0.0.17
-->::/0
|-->::/64
|-->2001:7c0:900:1c2::/64
||-->2001:7c0:900:1c2::/127
|||-->2001:7c0:900:1c2::1/128
`

	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	wantDepth := 3
	wantWidth := 4

	if maxDepth != wantDepth {
		t.Errorf("got max depth: %v want: %v", maxDepth, wantDepth)
	}
	if maxWidth != wantWidth {
		t.Errorf("got max width: %v want: %v", maxWidth, wantWidth)
	}
}

func TestTreeWalkStop(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/127",
		"2001:7c0:900:1c2::1/128",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	var walkFn WalkFunc = func(n *Node, depth int) error {
		if depth > 2 {
			return fmt.Errorf("stop")
		}
		return nil
	}

	err := tr.Walk(walkFn)

	if err == nil {
		t.Errorf("walkFunc with stop condition, expected error!")
	}
}

func TestTreeInsertOne(t *testing.T) {
	r1 := Item{Block: inet.MustBlock("0.0.0.0/0")}

	tr := New()
	tr.MustInsert(r1)

	got := new(strings.Builder)
	tr.Fprint(got)

	want := `▼
└─ 0.0.0.0/0
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeMultiRoot(t *testing.T) {
	r1 := Item{inet.MustBlock("0.0.0.0/0"), nil, nil}
	r2 := Item{inet.MustBlock("::/0"), nil, nil}

	tr := New()
	tr.MustInsert(r1)
	tr.MustInsert(r2)

	got := new(strings.Builder)
	tr.Fprint(got)

	want := `▼
├─ 0.0.0.0/0
└─ ::/0
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemoveEmpty(t *testing.T) {
	tr := New()
	r := Item{inet.MustBlock("0.0.0.0/0"), nil, nil}

	err := tr.Remove(r)
	if err == nil {
		t.Error("Remove() in empty tree, expected error!")
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemoveEdgeCase(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"10.0.0.0-10.0.0.30",
		"10.0.0.2-10.0.0.50",
		"10.0.0.20-10.0.0.25",
		"10.0.0.24-10.0.0.35",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r := Item{inet.MustBlock("10.0.0.2-10.0.0.50"), nil, nil}
	if err := tr.Remove(r); err != nil {
		t.Errorf("Remove(%v): %s\n", r, err)
	}

	w2 := new(strings.Builder)
	tr.Fprint(w2)

	// edge case, childs get resorted to different parents
	want :=
		`▼
├─ 10.0.0.0-10.0.0.30
│  └─ 10.0.0.20-10.0.0.25
└─ 10.0.0.24-10.0.0.35
`

	if w2.String() != want {
		t.Errorf("start:\n%s\nremove %v\ngot:\n%s\nwant:\n%s\n", w1.String(), r, w2, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemove(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/96",
		"2001:7c0:900:1c2::0/120",
		"2001:7c0:900:1c2::1/128",
		"2001:7c0:900:1c2::5/128",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r := Item{Block: inet.MustBlock("2001:7c0:900:1c2::/64")}
	if err := tr.Remove(r); err != nil {
		t.Errorf("Remove(%v): %s\n", r, err)
	}

	w2 := new(strings.Builder)
	tr.Fprint(w2)

	want :=
		`▼
├─ 0.0.0.0/0
│  ├─ 0.0.0.0/8
│  │  └─ 0.0.0.0/10
│  ├─ 1.0.0.0/8
│  ├─ 5.0.0.0/8
│  └─ 10.0.0.0-10.0.0.17
└─ ::/0
   ├─ ::/64
   └─ 2001:7c0:900:1c2::/96
      └─ 2001:7c0:900:1c2::/120
         ├─ 2001:7c0:900:1c2::1/128
         └─ 2001:7c0:900:1c2::5/128
`

	if w2.String() != want {
		t.Errorf("start:\n%s\nremove %v\ngot:\n%s\nwant:\n%s\n", w1.String(), r, w2, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemoveBranch(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/96",
		"2001:7c0:900:1c2::0/120",
		"2001:7c0:900:1c2::1/128",
		"2001:7c0:900:1c2::5/128",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r := Item{Block: inet.MustBlock("2001:7c0:900:1c2::/64")}
	if err := tr.RemoveBranch(r); err != nil {
		t.Errorf("RemoveBranch(%v): %s\n", r, err)
	}

	w2 := new(strings.Builder)
	tr.Fprint(w2)

	want :=
		`▼
├─ 0.0.0.0/0
│  ├─ 0.0.0.0/8
│  │  └─ 0.0.0.0/10
│  ├─ 1.0.0.0/8
│  ├─ 5.0.0.0/8
│  └─ 10.0.0.0-10.0.0.17
└─ ::/0
   └─ ::/64
`

	if w2.String() != want {
		t.Errorf("start:\n%s\nremove %v\ngot:\n%s\nwant:\n%s\n", w1.String(), r, w2, want)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemoveFalse(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/0",
		"1.0.0.0/8",
		"2.0.0.0/8",
		"4.0.0.0/8",
		"5.0.0.0/8",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	r := Item{inet.MustBlock("0.0.0.0/8"), nil, nil}
	if err := tr.Remove(r); err == nil {
		t.Errorf("Remove(%v): not in tree, expected error\n", r)
	}

	r = Item{inet.MustBlock("6.0.0.0/8"), nil, nil}
	if err := tr.Remove(r); err == nil {
		t.Errorf("Remove(%v): not in tree, expected error\n", r)
	}

	r = Item{inet.MustBlock("3.0.0.0/8"), nil, nil}
	if err := tr.Remove(r); err == nil {
		t.Errorf("Remove(%v): not in tree, expected error\n", r)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func TestTreeRemoveInsert(t *testing.T) {
	tr := New()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/127",
		"2001:7c0:900:1c2::1/128",
		"0.0.0.0/10",
	} {
		item := Item{inet.MustBlock(s), nil, nil}
		tr.MustInsert(item)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r := Item{Block: inet.MustBlock("2001:7c0:900:1c2::0/127")}

	// test idempotent
	if err := tr.Remove(r); err != nil {
		t.Errorf("Remove(%v): %s\n", r, err)
	}

	w2 := new(strings.Builder)
	tr.Fprint(w2)

	tr.MustInsert(r)

	w3 := new(strings.Builder)
	tr.Fprint(w3)

	if w1.String() != w3.String() {
		t.Errorf("remove and insert CIDR(%s) not idempotent:\n%s\n%s\n%s\n", r, w1, w2, w3)
	}

	if err := tr.Walk(nodeIsValid); err != nil {
		t.Errorf("tree checker returned: %s", err)
	}
}

func nodeIsValid(n *Node, d int) error {

	// check parent-child relationship
	if n.Parent != nil && n.Parent.Item != nil && n.Item != nil {
		if !n.Parent.Item.Block.Contains(n.Item.Block) {
			return fmt.Errorf("parent (%s) doesn't contain node (%s)", n.Parent.Item, n.Item)
		}
	}

	if len(n.Childs) == 0 {
		return nil
	}

	// check if node contains any child
	for i := range n.Childs {
		if !n.Item.Block.Contains(n.Childs[i].Item.Block) {
			return fmt.Errorf("node (%s) doesn't contain child (%s)", n.Item, n.Childs[i].Item)
		}
	}

	// check if childs are sorted
	less := func(i, j int) bool { return n.Childs[i].Item.Block.Compare(n.Childs[j].Item.Block) < 0 }

	if !sort.SliceIsSorted(n.Childs, less) {
		return fmt.Errorf("node: %s, childs aren't sorted", n.Item)
	}

	// check if the childs in row contains themselves
	for i := 0; i < len(n.Childs)-1; i++ {
		if n.Childs[i].Item.Block.Contains(n.Childs[i+1].Item.Block) {
			return fmt.Errorf("child (%s) contains child (%s) in row", n.Childs[i].Item, n.Childs[i+1].Item)
		}
	}

	// check for dup childs in row
	for i := 0; i < len(n.Childs)-1; i++ {
		if n.Childs[i].Item.Block.Compare(n.Childs[i+1].Item.Block) == 0 {
			return fmt.Errorf("child (%s) is dup in row", n.Childs[i].Item)
		}
	}

	return nil
}
