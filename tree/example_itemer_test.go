package tree_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func asString(b inet.Block, p interface{}) string {
	cidr := b.String()
	spaceDots := 38 - len(cidr)
	spacer := strings.Repeat(".", spaceDots)
	return fmt.Sprintf("%s %s %v", cidr, spacer, p)
}

func ExampleItemer_interface() {
	tr := tree.New()
	for _, r := range []tree.Item{
		tree.Item{inet.MustBlock("0.0.0.0/8"), "lorem", asString},
		tree.Item{inet.MustBlock("1.0.0.0/8"), "lorem", asString},
		tree.Item{inet.MustBlock("::/64"), "lorem", asString},
		tree.Item{inet.MustBlock("5.0.0.0/8"), "lorem", asString},
		tree.Item{inet.MustBlock("0.0.0.0/0"), "lorem", asString},
		tree.Item{inet.MustBlock("10.0.0.0-10.0.0.17"), "lorem", asString},
		tree.Item{inet.MustBlock("::/0"), "dup cidr", asString},
		tree.Item{inet.MustBlock("::/0"), "dup cidr", asString},
		tree.Item{inet.MustBlock("2001:7c0:900:1c2::/64"), "lorem", asString},
		tree.Item{inet.MustBlock("2001:7c0:900:1c2::0/127"), "lorem", asString},
		tree.Item{inet.MustBlock("2001:7c0:900:1c2::1/128"), "lorem", asString},
		tree.Item{inet.MustBlock("0.0.0.0/10"), "lorem", asString},
		// ...
	} {
		tr.Insert(r)
	}

	tr.Fprint(os.Stdout)

	// Output:
	// ▼
	// ├─ 0.0.0.0/0 ............................. lorem
	// │  ├─ 0.0.0.0/8 ............................. lorem
	// │  │  └─ 0.0.0.0/10 ............................ lorem
	// │  ├─ 1.0.0.0/8 ............................. lorem
	// │  ├─ 5.0.0.0/8 ............................. lorem
	// │  └─ 10.0.0.0-10.0.0.17 .................... lorem
	// └─ ::/0 .................................. dup cidr
	//    ├─ ::/64 ................................. lorem
	//    └─ 2001:7c0:900:1c2::/64 ................. lorem
	//       └─ 2001:7c0:900:1c2::/127 ................ lorem
	//          └─ 2001:7c0:900:1c2::1/128 ............... lorem

}
