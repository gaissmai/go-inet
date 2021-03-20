// An example tree.Interface implementation.
package tree_test

import (
	"fmt"
	"log"

	"github.com/gaissmai/go-inet/v2/tree"
)

// simplest interval type ever
type ival struct {
	lo, hi int
}

// --- Interface implementation
//
// Equal
func (a ival) Equal(i tree.Interface) bool {
	b := i.(ival)
	return a == b
}

// Covers
func (a ival) Covers(i tree.Interface) bool {
	b := i.(ival)
	if a == b {
		return false
	}
	return a.lo <= b.lo && a.hi >= b.hi
}

// Less
func (a ival) Less(i tree.Interface) bool {
	b := i.(ival)
	if a == b {
		return false
	}
	if a.lo == b.lo {
		// HINT: sort containers to the left!
		return a.hi > b.hi
	}
	return a.lo < b.lo
}

// fmt.Stringer
func (a ival) String() string {
	return fmt.Sprintf("%d...%d", a.lo, a.hi)
}

// ---

// global values for the examples
var is = []tree.Interface{
	ival{1, 10},
	ival{6, 40},
	ival{17, 17},
	ival{6, 200},
	ival{20, 42},
	ival{7, 8},
	ival{4, 6},
	ival{6, 33},
	ival{7, 39},
	ival{4, 12},
}

func Example_interface() {

	t, err := tree.NewTree(is)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t)
	fmt.Printf("Len:         %v\n", t.Len())
	fmt.Println()
	fmt.Printf("Lookup:      %v => %v\n", ival{7, 22}, t.Lookup(ival{7, 22}))
	fmt.Printf("Superset:    %v => %v\n", ival{7, 22}, t.Superset(ival{7, 22}))
	fmt.Println()
	fmt.Printf("Lookup:      %v => %v\n", ival{7, 7}, t.Lookup(ival{7, 7}))
	fmt.Printf("Superset:    %v => %v\n", ival{7, 7}, t.Superset(ival{7, 7}))

	// Output:
	// ▼
	// ├─ 1...10
	// ├─ 4...12
	// │  └─ 4...6
	// └─ 6...200
	//    ├─ 6...40
	//    │  ├─ 6...33
	//    │  └─ 7...39
	//    │     ├─ 7...8
	//    │     └─ 17...17
	//    └─ 20...42
	//
	// Len:         10
	//
	// Lookup:      7...22 => 7...39
	// Superset:    7...22 => 6...200
	//
	// Lookup:      7...7 => 7...8
	// Superset:    7...7 => 1...10
}
