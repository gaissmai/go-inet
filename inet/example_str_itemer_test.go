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
