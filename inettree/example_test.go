package inettree_test

import (
	"fmt"

	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/inettree"
	"github.com/gaissmai/go-inet/v2/tree"
)

var iana = map[string]string{
	"::/8":     "Reserved by IETF     [RFC3513][RFC4291]",
	"100::/8":  "Reserved by IETF     [RFC3513][RFC4291]",
	"200::/7":  "Reserved by IETF     [RFC4048]",
	"400::/6":  "Reserved by IETF     [RFC3513][RFC4291]",
	"800::/5":  "Reserved by IETF     [RFC3513][RFC4291]",
	"1000::/4": "Reserved by IETF     [RFC3513][RFC4291]",
	"2000::/3": "Global Unicast       [RFC3513][RFC4291]",
	"2000::/4": "",
	"3000::/4": "FREE",
	"4000::/3": "Reserved by IETF     [RFC3513][RFC4291]",
	"6000::/3": "Reserved by IETF     [RFC3513][RFC4291]",
}

func Example_usage() {
	bs := make([]tree.Interface, 0, len(iana))

	for k, text := range iana {
		block, err := inet.ParseBlock(k)
		if err != nil {
			panic(err)
		}
		if text != "" {
			text = block.String() + " ... " + text
		}
		bs = append(bs, inettree.Item{block, text})
	}

	t, err := tree.NewTree(bs)
	if err != nil {
		panic(err)
	}

	fmt.Println(t)

	// Output:
	// ▼
	// ├─ ::/8 ... Reserved by IETF     [RFC3513][RFC4291]
	// ├─ 100::/8 ... Reserved by IETF     [RFC3513][RFC4291]
	// ├─ 200::/7 ... Reserved by IETF     [RFC4048]
	// ├─ 400::/6 ... Reserved by IETF     [RFC3513][RFC4291]
	// ├─ 800::/5 ... Reserved by IETF     [RFC3513][RFC4291]
	// ├─ 1000::/4 ... Reserved by IETF     [RFC3513][RFC4291]
	// ├─ 2000::/3 ... Global Unicast       [RFC3513][RFC4291]
	// │  ├─ 2000::/4
	// │  └─ 3000::/4 ... FREE
	// ├─ 4000::/3 ... Reserved by IETF     [RFC3513][RFC4291]
	// └─ 6000::/3 ... Reserved by IETF     [RFC3513][RFC4291]

}
