package inet_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/inet"
)

func ExampleTree_Lookup() {
	bt := inet.NewTree()
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

func ExampleTree_Walk() {
	tr := inet.NewTree()

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
