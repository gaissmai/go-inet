package tree_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func ExampleTree_Lookup() {
	bs := make([]inet.Block, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		bs = append(bs, inet.MustBlock(s))
	}

	bt := tree.New()
	bt.InsertBulk(bs)

	q := inet.MustBlock(inet.MustIP("5.0.122.12"))

	if match, ok := bt.Lookup(q); ok {
		fmt.Printf("inet.Lookup(%v): LPM found at: %v\n", q, match)
	}

	// Output:
	// inet.Lookup(5.0.122.12/32): LPM found at: 5.0.0.0/8

}

func ExampleTree_Lookup_exists() {
	bs := make([]inet.Block, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		bs = append(bs, inet.MustBlock(s))
	}

	bt := tree.New()
	bt.InsertBulk(bs)

	// look for exists, exact match, not just LPM
	for _, s := range []string{
		"5.0.0.0/8",
		"5.0.1.2/32",
	} {
		q := inet.MustBlock(s)

		match, ok := bt.Lookup(q)
		if ok && match.Compare(q) == 0 {
			fmt.Printf("%q exists in tree\n", q)
		} else {
			fmt.Printf("%q doesn't exists in tree\n", q)
		}
	}

	// Output:
	// "5.0.0.0/8" exists in tree
	// "5.0.1.2/32" doesn't exists in tree

}

func ExampleTree_Walk() {

	bs := make([]inet.Block, 0)
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
		bs = append(bs, inet.MustBlock(s))
	}
	tr := tree.New()
	tr.InsertBulk(bs)

	var maxDepth int
	var maxWidth int

	var walkFn tree.WalkFunc = func(n *tree.Node, depth int) error {
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
