package tree

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gaissmai/go-inet/inet"
)

func TestTreeLookup(t *testing.T) {
	s1, _ := inet.NewBlock("0.0.0.0/0")

	tree := NewBlockTree().Insert(s1)

	got, ok := tree.Lookup(s1)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", s1, ok, true)
	}

	if got.Compare(s1) != 0 {
		t.Errorf("tree.Lookup(%v), got: %v, want: %v", s1, got, s1)
	}
}

func TestTreeLookupLPM(t *testing.T) {
	tr := NewBlockTree()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		item := inet.MustBlock(inet.NewBlock(s))
		tr.Insert(item)
	}

	look := inet.MustIP(inet.NewIP("0.0.0.0"))
	want := inet.MustBlock(inet.NewBlock("0.0.0.0/10"))

	got, ok := tr.Lookup(look)
	if !ok {
		t.Errorf("Lookup(%s) got %t, want %t", look, ok, true)
	}

	if got.Compare(want) != 0 {
		t.Errorf("tree.Lookup(%v), got: %v, want: %v", look, got, want)
	}
}

func TestTreeWalk(t *testing.T) {
	tr := NewBlockTree()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		item := inet.MustBlock(inet.NewBlock(s))
		tr.Insert(item)
	}

	got := new(strings.Builder)

	var walkFn WalkFunc = func(n *Node, depth int) error {
		pfx := strings.Repeat("|", depth)
		fmt.Fprintf(got, "%s-->%s\n", pfx, *n.Item)
		return nil
	}

	err := tr.Walk(walkFn, true)

	if err != nil {
		t.Errorf("%v\n", err)
	}

	want :=
		`-->0.0.0.0/0
|-->0.0.0.0/8
||-->0.0.0.0/10
|-->1.0.0.0/8
|-->5.0.0.0/8
-->::/0
|-->::/64
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}
}

func TestTreeInsertDup(t *testing.T) {
	r1, _ := inet.NewBlock("0.0.0.0/0")

	tree := NewBlockTree().InsertBulk(r1, r2)

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
	r1, _ := inet.NewBlock("0.0.0.0/0")
	r2, _ := inet.NewBlock("::/0")

	tree := NewBlockTree().InsertBulk(r1, r2)

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

func TestTreeRemove(t *testing.T) {
	s1, _ := inet.NewBlock("10.0.0.0/0")
	s2, _ := inet.NewBlock("10.0.0.0/4")
	s3, _ := inet.NewBlock("10.0.0.0/8")

	tree := NewBlockTree().InsertBulk(s1, s2, s3)

	got := new(strings.Builder)
	tree.Fprint(got)

	want := `▼
└─ 0.0.0.0/0
   └─ 0.0.0.0/4
      └─ 10.0.0.0/8
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if ok := tree.Remove(s2); !ok {
		t.Errorf("tree.Remove(%v), got: %v, want: %v", s2, ok, true)
	}

	got.Reset()
	tree.Fprint(got)

	want = `▼
└─ 0.0.0.0/0
   └─ 10.0.0.0/8
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if ok := tree.Remove(s1); !ok {
		t.Errorf("tree.Remove(%v), got: %v, want: %v", s1, ok, true)
	}

	got.Reset()
	tree.Fprint(got)

	want = `▼
└─ 10.0.0.0/8
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

	if ok := tree.Remove(s3); !ok {
		t.Errorf("tree.Remove(%v), got: %v, want: %v", s3, ok, true)
	}

	got.Reset()
	tree.Fprint(got)

	want = `▼
`
	if got.String() != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}

}
