package tree_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func ExampleTree_Lookup() {
	is := make([]tree.Item, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		is = append(is, tree.Item{Block: inet.MustBlock(s)})
	}

	tr := tree.New()
	tr.InsertBulk(is)

	q := tree.Item{Block: inet.MustBlock(inet.MustIP("5.0.122.12"))}

	if match, ok := tr.Lookup(q); ok {
		fmt.Printf("tree.Lookup(%v): LPM found at: %v\n", q, match)
	}

	// Output:
	// tree.Lookup(5.0.122.12/32): LPM found at: 5.0.0.0/8

}

func ExampleTree_Lookup_exists() {
	is := make([]tree.Item, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"::/64",
		"::/0",
		"0.0.0.0/10",
	} {
		is = append(is, tree.Item{Block: inet.MustBlock(s)})
	}

	tr := tree.New()
	tr.InsertBulk(is)

	// look for exists, exact match, not just LPM
	for _, s := range []string{
		"5.0.0.0/8",
		"5.0.1.2/32",
	} {
		q := tree.Item{Block: inet.MustBlock(s)}

		match, ok := tr.Lookup(q)
		if ok && match.Block.Compare(q.Block) == 0 {
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

	is := make([]tree.Item, 0)
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
		is = append(is, tree.Item{Block: inet.MustBlock(s)})
	}
	tr := tree.New()
	tr.InsertBulk(is)

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