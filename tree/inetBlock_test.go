package tree_test

import (
	"strings"
	"testing"

	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/tree"
)

type inetItem struct {
	inet.Block
	text string
}

func NewInetItem(b string, ss ...string) inetItem {
	bb, err := inet.ParseBlock(b)
	if err != nil {
		panic(err)
	}
	return inetItem{Block: bb, text: strings.Join(ss, " ")}
}

// ###########################################
// implement the Item interface for inet.Block
// ###########################################

func (b inetItem) Less(i tree.Interface) bool {
	if b.Block == i.(inetItem).Block {
		return b.text < i.(inetItem).text
	}

	return b.Block.Less(i.(inetItem).Block)
}

func (b inetItem) Equal(i tree.Interface) bool {
	return b == i.(inetItem)
}

func (b inetItem) Covers(i tree.Interface) bool {
	return b.Block.Covers(i.(inetItem).Block)
}

func (b inetItem) String() string {
	if b.text == "" {
		return b.Block.String()
	}
	return b.Block.String() + " > " + b.text
}

// #####################################################

func TestIntervalTreeInsert(t *testing.T) {

	is := []tree.Interface{
		NewInetItem("2001:db8::/32"),
		NewInetItem("127.0.0.1"),
		NewInetItem("::/0"),
		NewInetItem("::1"),
		NewInetItem("0.0.0.0/0"),
	}

	tr, _ := tree.NewTree(is)

	got := tr.String()
	want := `▼
├─ 0.0.0.0/0
│  └─ 127.0.0.1/32
└─ ::/0
   ├─ ::1/128
   └─ 2001:db8::/32
`

	if got != want {
		t.Errorf("got:\n%v\nwant:\n%v", got, want)
	}
}

func TestIntervalTreeLookup(t *testing.T) {
	it := []tree.Interface{
		NewInetItem("0.0.0.0/0"),
		NewInetItem("127.0.0.0/8"),
		NewInetItem("192.168.0.0/16"),
		NewInetItem("::/0"),
		NewInetItem("fe80::/10"),
		NewInetItem("2001:db8::/32"),
	}

	tr, _ := tree.NewTree(it)

	tests := []struct {
		in   inetItem
		want inetItem
	}{
		{NewInetItem("0.0.0.0/0"), NewInetItem("0.0.0.0/0")},
		{NewInetItem("::/0"), NewInetItem("::/0")},
		{NewInetItem("127.0.0.1"), NewInetItem("127.0.0.0/8")},
		{NewInetItem("192.168.0.1"), NewInetItem("192.168.0.0/16")},
		{NewInetItem("2001:db8::affe"), NewInetItem("2001:db8::/32")},
		{NewInetItem("::1"), NewInetItem("::/0")},
		{NewInetItem("fe80::/12"), NewInetItem("fe80::/10")},
	}

	for _, tt := range tests {
		got := tr.Lookup(tt.in)
		if got != tt.want {
			t.Errorf("Lookup(%v) = %v; want %v", tt.in, got, tt.want)
		}
	}

	if got := tr.Lookup(nil); got != nil {
		t.Errorf("Lookup(nil) = %v; want %v", got, nil)
	}
}

func TestIntervalTreeSuperset(t *testing.T) {
	it := []tree.Interface{
		NewInetItem("0.0.0.0/4"),
		NewInetItem("127.0.0.0/8"),
		NewInetItem("192.168.0.0/16"),
		NewInetItem("192.169.0.0/16"),
		NewInetItem("192.170.0.0/16"),
		NewInetItem("192.171.0.0/16"),
		NewInetItem("::/4"),
		NewInetItem("fe80::/10"),
		NewInetItem("2001:db8::/32"),
	}

	tr, _ := tree.NewTree(it)

	valid := []struct {
		in   inetItem
		want inetItem
	}{
		{NewInetItem("0.0.0.0/6"), NewInetItem("0.0.0.0/4")},
		{NewInetItem("::/6"), NewInetItem("::/4")},
		{NewInetItem("127.0.0.1"), NewInetItem("127.0.0.0/8")},
		{NewInetItem("192.168.0.1"), NewInetItem("192.168.0.0/16")},
		{NewInetItem("192.169.0.0/16"), NewInetItem("192.169.0.0/16")},
		{NewInetItem("2001:db8::affe"), NewInetItem("2001:db8::/32")},
		{NewInetItem("::1"), NewInetItem("::/4")},
		{NewInetItem("fe80::/12"), NewInetItem("fe80::/10")},
	}

	for _, tt := range valid {
		got := tr.Superset(tt.in)
		if got != tt.want {
			t.Errorf("Superset(%v) = %v; want %v", tt.in, got, tt.want)
		}
	}

	invalid := []struct {
		in inetItem
	}{
		{NewInetItem("17.0.0.0/4")},
		{NewInetItem("::/3")},
		{NewInetItem("128.0.0.1")},
		{NewInetItem("168.192.0.1")},
		{NewInetItem("2001:db9::affe")},
		{NewInetItem("::/0")},
		{NewInetItem("fe80::/9")},
	}

	for _, tt := range invalid {
		got := tr.Superset(tt.in)
		if got != nil {
			t.Errorf("Superset(%v) =  %v; want: %v", tt.in, got, nil)
		}
	}

	if got := tr.Superset(nil); got != nil {
		t.Errorf("Superset(nil) = %v; want %v", got, nil)
	}
}

func TestIntervalTreeLookupMissing(t *testing.T) {
	it := []tree.Interface{
		NewInetItem("2001:db8::/32"),
		NewInetItem("::/0"),
		NewInetItem("::1"),
	}

	tr, _ := tree.NewTree(it)

	not := NewInetItem("1.2.3.4/5")
	if got := tr.Lookup(not); got != nil {
		t.Errorf("Lookup(%v) = %v; want: %v", not, got, nil)
	}
}
