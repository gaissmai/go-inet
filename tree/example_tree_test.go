package tree_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func ExampleBlockTree_Lookup() {
	bt := tree.NewBlockTree()
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

	null := inet.MustBlock(inet.MustIP("0.0.0.0"))

	if match, ok := bt.Lookup(null); ok {
		fmt.Printf("tree.Lookup('0.0.0.0'): LPM found at: %v\n", match)
	}

	// Output:
	// tree.Lookup('0.0.0.0'): LPM found at: 0.0.0.0/10

}
