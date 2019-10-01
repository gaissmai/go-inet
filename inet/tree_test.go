package inet

import (
	"fmt"
	"strings"
	"testing"
)

func TestTreeLookupMissing(t *testing.T) {
	s1 := MustBlock("0.0.0.0/0")
	s2 := MustBlock("2001:db8::/32")

	tree := NewTree().Insert(s1)

	_, ok := tree.Lookup(s2)
	if ok {
		t.Errorf("Lookup(%s) got %t, want %t", s2, ok, false)
	}
}

func TestTreeLookup(t *testing.T) {
	s1 := MustBlock("0.0.0.0/0")

	tree := NewTree().Insert(s1)

	got, ok := tree.Lookup(s1)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", s1, ok, true)
	}

	if got.Compare(s1) != 0 {
		t.Errorf("tree.Lookup(%v), got: %v, want: %v", s1, got, s1)
	}
}

func TestTreeLookupLPM(t *testing.T) {
	tr := NewTree()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		item := MustBlock(s)
		tr.Insert(item)
	}

	look := MustBlock(MustIP("0.0.0.0"))
	want := MustBlock("0.0.0.0/10")

	got, ok := tr.Lookup(look)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", look, ok, true)
	}

	if got.Compare(want) != 0 {
		t.Errorf("tree.Lookup(%v), got: %v, want: %v", look, got, want)
	}
}

func TestTreeWalk(t *testing.T) {
	tr := NewTree()

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
		item := MustBlock(s)
		tr.Insert(item)
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

func TestTreeInsertDup(t *testing.T) {
	r1, _ := NewBlock("0.0.0.0/0")

	tree := NewTree()
	tree.Insert(r1)
	tree.Insert(r1)

	got := new(strings.Builder)
	tree.Fprint(got)

	want := `▼
└─ 0.0.0.0/0
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}
}

func TestTreeMultiRoot(t *testing.T) {
	r1, _ := NewBlock("0.0.0.0/0")
	r2, _ := NewBlock("::/0")

	tree := NewTree()
	tree.Insert(r1)
	tree.Insert(r2)

	got := new(strings.Builder)
	tree.Fprint(got)

	want := `▼
├─ 0.0.0.0/0
└─ ::/0
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}
}

func TestTreeRemoveEdgeCase(t *testing.T) {
	tr := NewTree()

	for _, s := range []string{
		"10.0.0.0-10.0.0.30",
		"10.0.0.2-10.0.0.50",
		"10.0.0.20-10.0.0.25",
		"10.0.0.24-10.0.0.35",
	} {
		item := MustBlock(s)
		tr.Insert(item)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r := MustBlock("10.0.0.2-10.0.0.50")
	got := tr.Remove(r)
	if !got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, true)
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
}

func TestTreeRemove(t *testing.T) {
	tr := NewTree()

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
		item := MustBlock(s)
		tr.Insert(item)
	}

	r := MustBlock("3.0.0.0/8")
	got := tr.Remove(r)
	if got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, false)
	}

	w1 := new(strings.Builder)
	tr.Fprint(w1)

	r = MustBlock("2001:7c0:900:1c2::0/127")
	got = tr.Remove(r)
	if !got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, true)
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
   └─ 2001:7c0:900:1c2::/64
      └─ 2001:7c0:900:1c2::1/128
`

	if w2.String() != want {
		t.Errorf("start:\n%s\nremove %v\ngot:\n%s\nwant:\n%s\n", w1.String(), r, w2, want)
	}
}

func TestTreeRemoveFalse(t *testing.T) {
	tr := NewTree()

	for _, s := range []string{
		"0.0.0.0/0",
		"1.0.0.0/8",
		"2.0.0.0/8",
		"4.0.0.0/8",
		"5.0.0.0/8",
	} {
		item := MustBlock(s)
		tr.Insert(item)
	}

	// frist pos in childs
	r := MustBlock("0.0.0.0/8")
	got := tr.Remove(r)
	if got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, false)
	}

	// last pos in childs
	r = MustBlock("6.0.0.0/8")
	got = tr.Remove(r)
	if got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, false)
	}

	// middle pos in childs
	r = MustBlock("6.0.0.0/8")
	got = tr.Remove(r)
	if got {
		t.Errorf("Remove(%v), got %t, want %t\n", r, got, false)
	}
}