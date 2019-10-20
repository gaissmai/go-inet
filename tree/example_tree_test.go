package tree_test

import (
	"fmt"
	"os"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func ExampleTree_Insert() {

	// String callback function used in tree.Fprintf()
	callback := func(i tree.Item) string {
		return fmt.Sprintf("%s %s %v", i.Block, ".........", i.Payload)
	}

	items := []tree.Item{
		{inet.MustBlock("0.0.0.0/8"), "text as payload", callback},
		{inet.MustBlock("1.0.0.0/8"), "text as payload", callback},
		{inet.MustBlock("::/64"), "text as payload", callback},
		{inet.MustBlock("5.0.0.0/8"), "text as payload", callback},
		{inet.MustBlock("0.0.0.0/0"), "text as payload", callback},
		{inet.MustBlock("10.0.0.0-10.0.0.17"), "text as payload", callback},
		{inet.MustBlock("::/0"), "text as payload", callback},
		{inet.MustBlock("2001:7c0:900:1c2::/64"), "text as payload", callback},
		{inet.MustBlock("2001:7c0:900:1c2::0/127"), "text as payload", callback},
		{inet.MustBlock("2001:7c0:900:1c2::1/128"), "text as payload", callback},
		{inet.MustBlock("0.0.0.0/10"), "text as payload", callback},
		// ...
	}

	tr := tree.New()
	if err := tr.Insert(items...); err != nil {
		panic(err)
	}
	tr.Fprint(os.Stdout)

	// Output:
	// ▼
	// ├─ 0.0.0.0/0 ......... text as payload
	// │  ├─ 0.0.0.0/8 ......... text as payload
	// │  │  └─ 0.0.0.0/10 ......... text as payload
	// │  ├─ 1.0.0.0/8 ......... text as payload
	// │  ├─ 5.0.0.0/8 ......... text as payload
	// │  └─ 10.0.0.0-10.0.0.17 ......... text as payload
	// └─ ::/0 ......... text as payload
	//    ├─ ::/64 ......... text as payload
	//    └─ 2001:7c0:900:1c2::/64 ......... text as payload
	//       └─ 2001:7c0:900:1c2::/127 ......... text as payload
	//          └─ 2001:7c0:900:1c2::1/128 ......... text as payload

}

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
	tr.MustInsert(is...)

	q := tree.Item{Block: inet.MustBlock(inet.MustIP("5.0.122.12"))}

	if match, ok := tr.Lookup(q); ok {
		fmt.Printf("tree.Lookup(%v): LPM found at: %v\n", q, match)
	}

	// Output:
	// tree.Lookup(5.0.122.12/32): LPM found at: 5.0.0.0/8

}

func ExampleTree_Contains() {
	is := make([]tree.Item, 0)
	for _, s := range []string{
		"1.0.0.0/8",
		"5.0.0.0/8",
		"::/64",
		"0.0.0.0/10",
	} {
		is = append(is, tree.Item{Block: inet.MustBlock(s)})
	}

	tr := tree.New()
	tr.MustInsert(is...)

	// look for containment in tree
	for _, s := range []string{
		"0.0.0.2/8",
		"5.0.1.2/32",
		"2001:db8::1/128",
	} {
		q := tree.Item{Block: inet.MustBlock(s)}

		ok := tr.Contains(q)
		if ok {
			fmt.Printf("%-15s is    contained in tree\n", q)
		} else {
			fmt.Printf("%-15s isn't contained in tree\n", q)
		}
	}

	// Output:
	// 0.0.0.0/8       isn't contained in tree
	// 5.0.1.2/32      is    contained in tree
	// 2001:db8::1/128 isn't contained in tree

}

func ExampleTree_Walk() {

	is := make([]tree.Item, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"5.0.0.0/8",
		"6.0.0.0/8",
		"7.0.0.0/8",
		"8.0.0.0/8",
		"9.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:db8:900:1c2::/64",
		"2001:db8:900:1c2::0/127",
		"2001:db8:900:1c2::1/128",
	} {
		is = append(is, tree.Item{Block: inet.MustBlock(s)})
	}
	tr := tree.New()
	tr.MustInsert(is...)

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
	tr.Fprint(os.Stdout)

	// Output:
	// max depth: 3
	// max width: 7
	// ▼
	// ├─ 0.0.0.0/0
	// │  ├─ 0.0.0.0/8
	// │  ├─ 5.0.0.0/8
	// │  ├─ 6.0.0.0/8
	// │  ├─ 7.0.0.0/8
	// │  ├─ 8.0.0.0/8
	// │  ├─ 9.0.0.0/8
	// │  └─ 10.0.0.0-10.0.0.17
	// └─ ::/0
	//    ├─ ::/64
	//    └─ 2001:db8:900:1c2::/64
	//       └─ 2001:db8:900:1c2::/127
	//          └─ 2001:db8:900:1c2::1/128

}
