package inet_test

import (
	"fmt"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/inet"
)

type trieStr string

func (a trieStr) Contains(b inet.Itemer) bool {
	c, ok := b.(trieStr)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return strings.Index(string(c), string(a)) == 0
}

func (a trieStr) Compare(b inet.Itemer) int {
	c, ok := b.(trieStr)
	if !ok {
		panic(fmt.Errorf("incompatible types: %T != %T", a, b))
	}
	return strings.Compare(string(a), string(c))
}

func ExampleItemer() {
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
		tr.Insert(trieStr(s))
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
