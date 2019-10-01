package inet_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/inet"
)

type stringItem string

func (a stringItem) Contains(b inet.Itemer) bool {
	c, ok := b.(stringItem)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	// e.g. a = "foo", b=c="foobar"; a contains b in the sense of a trie
	return strings.Index(string(c), string(a)) == 0
}

func (a stringItem) Compare(b inet.Itemer) int {
	c, ok := b.(stringItem)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return strings.Compare(string(a), string(c))
}

func ExampleItemer_stringItem() {
	tr := inet.NewTrie()

	for _, s := range []string{
		"ha",
		"hallo",
		"f",
		"ha", // duplicate
		"barber shop",
		"hallo", // duplicate
		"foo_bar",
		"bar",
		"foo",
		"barbara",
		"fo",
	} {
		tr.Insert(stringItem(s))
	}

	tr.Fprint(os.Stdout)

	// Output:
	// ▼
	// ├─ bar
	// │  ├─ barbara
	// │  └─ barber shop
	// ├─ f
	// │  └─ fo
	// │     └─ foo
	// │        └─ foo_bar
	// └─ ha
	//    └─ hallo

}

func ExampleItemer_blockItem() {
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
