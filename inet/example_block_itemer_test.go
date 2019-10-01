package inet_test

import (
	"os"
	"github.com/gaissmai/go-inet/inet"
)

func ExampleItemer_Block() {
	bs := make([]inet.Block, 0)
	for _, s := range []string{
		"0.0.0.0/8",
		"1.0.0.0/8",
		"5.0.0.0/8",
		"0.0.0.0/0",
		"10.0.0.0-10.0.0.17",
		"::/64",
		"::/0",
		"2001:7c0:900:1c2::/64",
		"2001:7c0:900:1c2::0/127",
		"2001:7c0:900:1c2::1/128",
		"0.0.0.0/10",
		// ... 1_000_000 more blocks
	} {
		bs = append(bs, inet.MustBlock(s))
	}
	// sort before inserting makes it faster for lot of items
	inet.SortBlock(bs)

	tr := inet.NewTrie()
	for _, b := range bs {
		tr.Insert(b)
	}

	tr.Fprint(os.Stdout)

	// Output:
	// ▼
	// ├─ 0.0.0.0/0
	// │  ├─ 0.0.0.0/8
	// │  │  └─ 0.0.0.0/10
	// │  ├─ 1.0.0.0/8
	// │  ├─ 5.0.0.0/8
	// │  └─ 10.0.0.0-10.0.0.17
	// └─ ::/0
	//    ├─ ::/64
	//    └─ 2001:7c0:900:1c2::/64
	//       └─ 2001:7c0:900:1c2::/127
	//          └─ 2001:7c0:900:1c2::1/128


}
