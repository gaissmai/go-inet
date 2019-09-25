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
		item := tree.NewSimpleItem(inet.MustBlock(inet.NewBlock(s)))
		bt.Insert(item)
	}

	null := inet.MustIP(inet.NewIP("0.0.0.0"))
	item := tree.NewSimpleItem(null)

	if match, ok := bt.Lookup(item); ok {
		fmt.Printf("tree.Lookup('0.0.0.0'): LPM found at: %v\n", match)
	}

	// Output:
	// tree.Lookup('0.0.0.0'): LPM found at: 0.0.0.0/10

}
