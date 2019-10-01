package inet_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleTrie_Lookup() {
	bt := inet.NewTrie()
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		item := inet.MustBlock(s)
		bt.Insert(item)
	}

	q := inet.MustBlock(inet.MustIP("5.0.122.12"))

	if match, ok := bt.Lookup(q); ok {
		fmt.Printf("inet.Lookup(%v): LPM found at: %v\n", q, match)
	}

	// Output:
	// inet.Lookup(5.0.122.12/32): LPM found at: 5.0.0.0/8

}

func ExampleTrie_Walk() {
	tr := inet.NewTrie()

	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:db8:900:1c2::/64",
		"2001:db8:900:1c2::0/127",
		"2001:db8:900:1c2::1/128",
		"0.0.0.0/10",
	} {
		item := inet.MustBlock(s)
		tr.Insert(item)
	}

	var maxDepth int
	var maxWidth int

	var walkFn inet.WalkFunc = func(n *inet.Node, depth int) error {
		if depth > maxDepth {
			maxDepth = depth
		}
		if l := len(n.Childs); l > maxWidth {
			maxWidth = l
		}
		return nil
	}

	_ = tr.Walk(walkFn)

	fmt.Printf("max depth: %v\n", maxDepth)
	fmt.Printf("max width: %v\n", maxWidth)

	// Output:
	// max depth: 3
	// max width: 4

}

type trieStr string

func (a trieStr) Contains(b inet.Itemer) bool {
	c, ok := b.(trieStr)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return strings.Index(string(c), string(a)) == 0
}

func (a trieStr) Compare(b inet.Itemer) int {
	c, ok := b.(trieStr)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return strings.Compare(string(a), string(c))
}

func ExampleTrie_Strings() {
	tr := inet.NewTrie()

	for _, s := range []string{
		"ha",
		"hallo",
		"f",
		"ha", // duplicate
		"barber shop",
		"hallo", // duplicate
		"foo_bar",
		"bar",
		"foo",
		"barbara",
		"fo",
	} {
		tr.Insert(trieStr(s))
	}

	tr.Fprint(os.Stdout)

	// Output:
	// ▼
	// ├─ bar
	// │  ├─ barbara
	// │  └─ barber shop
	// ├─ f
	// │  └─ fo
	// │     └─ foo
	// │        └─ foo_bar
	// └─ ha
	//    └─ hallo


}
