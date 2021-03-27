package tree_test

import (
	"fmt"
	"log"

	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/tree"
)

// augment inet.Block
type item struct {
	inet.Block
}

func mustNewItem(b string) item {
	bb, err := inet.ParseBlock(b)
	if err != nil {
		panic(err)
	}
	return item{bb}
}

// ###########################################
// implement the tree.Interface for inet.Block
// ###########################################

func (a item) Less(i tree.Interface) bool {
	b := i.(item)
	return a.Block.Less(b.Block)
}

func (a item) Equals(i tree.Interface) bool {
	b := i.(item)
	return a.Block == b.Block
}

func (a item) Covers(i tree.Interface) bool {
	b := i.(item)
	return a.Block.Covers(b.Block)
}

func (a item) String() string {
	return a.Block.String()
}

// #####################################################

var is = []tree.Interface{
	mustNewItem("2001:db8::/32"),
	mustNewItem("2001:db8:1:fffe::/48"),
	mustNewItem("2001:db8:1:fffe::0-2001:db8:1:fffe::7a"),
	mustNewItem("2001:db8:2:fffe::4"),
	mustNewItem("2001:db8:2:fffe::5"),
	mustNewItem("2001:db8:2::/48"),
	mustNewItem("2001:db8:2:fffe::6"),
	mustNewItem("2001:db8:1:fffe::7"),
	mustNewItem("127.0.0.1"),
	mustNewItem("127.0.0.0/8"),
	mustNewItem("::1"),
	mustNewItem("10.0.0.16-10.1.3.254"),
	mustNewItem("10.0.0.233"),
	mustNewItem("fe80::/10"),
	mustNewItem("169.254.0.0/16"),
}

func Example_interface() {

	t, err := tree.NewTree(is)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t)
	fmt.Printf("Len:         %v\n", t.Len())
	fmt.Println()

	b := mustNewItem("2001:db8:1:fffe::2-2001:db8:1:fffe::3")
	fmt.Printf("Lookup:      %v => %v\n", b, t.Lookup(b))
	fmt.Printf("Superset:    %v => %v\n", b, t.Superset(b))
	fmt.Println()

	b = mustNewItem("fc00::1")
	fmt.Printf("Lookup:      %v => %v\n", b, t.Lookup(b))
	fmt.Printf("Superset:    %v => %v\n", b, t.Superset(b))

	// Output:
	// ▼
	// ├─ 10.0.0.16-10.1.3.254
	// │  └─ 10.0.0.233/32
	// ├─ 127.0.0.0/8
	// │  └─ 127.0.0.1/32
	// ├─ 169.254.0.0/16
	// ├─ ::1/128
	// ├─ 2001:db8::/32
	// │  ├─ 2001:db8:1::/48
	// │  │  └─ 2001:db8:1:fffe::-2001:db8:1:fffe::7a
	// │  │     └─ 2001:db8:1:fffe::7/128
	// │  └─ 2001:db8:2::/48
	// │     ├─ 2001:db8:2:fffe::4/128
	// │     ├─ 2001:db8:2:fffe::5/128
	// │     └─ 2001:db8:2:fffe::6/128
	// └─ fe80::/10
	//
	// Len:         15
	//
	// Lookup:      2001:db8:1:fffe::2/127 => 2001:db8:1:fffe::-2001:db8:1:fffe::7a
	// Superset:    2001:db8:1:fffe::2/127 => 2001:db8::/32
	//
	// Lookup:      fc00::1/128 => <nil>
	// Superset:    fc00::1/128 => <nil>
}
