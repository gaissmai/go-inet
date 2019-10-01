package inet_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/inet"
)

// record implements inet.Itemer
type record struct {
	cidr inet.Block
	txt  string
}

// Contains, part of the Itemer interface
func (a record) Contains(b inet.Itemer) bool {
	c, ok := b.(record)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return a.cidr.Contains(c.cidr) ||
		(a.cidr.Compare(c.cidr) == 0 && strings.Compare(a.txt, c.txt) < 0)
}

// Compare, part of the Itemer interface
func (a record) Compare(b inet.Itemer) int {
	c, ok := b.(record)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}

	cmp := a.cidr.Compare(c.cidr)
	if cmp != 0 {
		return cmp
	}
	return strings.Compare(a.txt, c.txt)
}

// String implements the Stringer interface
func (a record) String() string {
	cidr := a.cidr.String()
	spaceDots := 38 - len(cidr)
	spacer := strings.Repeat(".", spaceDots)
	return fmt.Sprintf("%s %s %s", cidr, spacer, a.txt)
}

func ExampleItemer_interface() {
	tr := inet.NewTrie()
	for _, r := range []record{
		record{inet.MustBlock("0.0.0.0/8"), "lorem"},
		record{inet.MustBlock("1.0.0.0/8"), "lorem"},
		record{inet.MustBlock("::/64"), "lorem"},
		record{inet.MustBlock("5.0.0.0/8"), "lorem"},
		record{inet.MustBlock("0.0.0.0/0"), "lorem"},
		record{inet.MustBlock("10.0.0.0-10.0.0.17"), "lorem"},
		record{inet.MustBlock("::/0"), "same cidr"},
		record{inet.MustBlock("::/0"), "same cidr, different text"},
		record{inet.MustBlock("2001:7c0:900:1c2::/64"), "lorem"},
		record{inet.MustBlock("2001:7c0:900:1c2::0/127"), "lorem"},
		record{inet.MustBlock("2001:7c0:900:1c2::1/128"), "lorem"},
		record{inet.MustBlock("0.0.0.0/10"), "lorem"},
		// ...
	} {
		tr.Insert(r)
	}

	tr.Fprint(os.Stdout)

/*

▼
├─ 0.0.0.0/0 ............................. lorem
│  ├─ 0.0.0.0/8 ............................. lorem
│  │  └─ 0.0.0.0/10 ............................ lorem
│  ├─ 1.0.0.0/8 ............................. lorem
│  ├─ 5.0.0.0/8 ............................. lorem
│  └─ 10.0.0.0-10.0.0.17 .................... lorem
└─ ::/0 .................................. same cidr
   └─ ::/0 .................................. same cidr, different text
      ├─ ::/64 ................................. lorem
      └─ 2001:7c0:900:1c2::/64 ................. lorem
         └─ 2001:7c0:900:1c2::/127 ................ lorem
            └─ 2001:7c0:900:1c2::1/128 ............... lorem

*/

}
