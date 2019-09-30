package inet_test

import (
	"fmt"

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
