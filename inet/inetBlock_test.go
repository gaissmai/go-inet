package inet_test

import (
	"strings"
	"testing"

	"github.com/gaissmai/go-inet/v2/inet"
	"github.com/gaissmai/go-inet/v2/tree"
)

type IPItem struct {
	inet.Block
	text string
}

func NewIPItem(b string, ss ...string) IPItem {
	bb, err := inet.ParseBlock(b)
	if err != nil {
		panic(err)
	}
	return IPItem{Block: bb, text: strings.Join(ss, " ")}
}

// ###########################################
// implement the Item interface for inet.Block
// ###########################################

func (b IPItem) Less(i tree.Interface) bool {
	if b.Block == i.(IPItem).Block {
		return b.text < i.(IPItem).text
	}

	return b.Block.Less(i.(IPItem).Block)
}

func (b IPItem) Equal(i tree.Interface) bool {
	return b == i.(IPItem)
}

func (b IPItem) Covers(i tree.Interface) bool {
	return b.Block.Covers(i.(IPItem).Block)
}

func (b IPItem) String() string {
	if b.text == "" {
		return b.Block.String()
	}
	return b.Block.String() + " > " + b.text
}

// #####################################################

func TestIntervalTreeInsert(t *testing.T) {

	is := []tree.Interface{
		NewIPItem("2001:db8::/32"),
		NewIPItem("127.0.0.1"),
		NewIPItem("::/0"),
		NewIPItem("::1"),
		NewIPItem("0.0.0.0/0"),
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
		NewIPItem("0.0.0.0/0"),
		NewIPItem("127.0.0.0/8"),
		NewIPItem("192.168.0.0/16"),
		NewIPItem("::/0"),
		NewIPItem("fe80::/10"),
		NewIPItem("2001:db8::/32"),
	}

	tr, _ := tree.NewTree(it)

	tests := []struct {
		in   IPItem
		want IPItem
	}{
		{NewIPItem("0.0.0.0/0"), NewIPItem("0.0.0.0/0")},
		{NewIPItem("::/0"), NewIPItem("::/0")},
		{NewIPItem("127.0.0.1"), NewIPItem("127.0.0.0/8")},
		{NewIPItem("192.168.0.1"), NewIPItem("192.168.0.0/16")},
		{NewIPItem("2001:db8::affe"), NewIPItem("2001:db8::/32")},
		{NewIPItem("::1"), NewIPItem("::/0")},
		{NewIPItem("fe80::/12"), NewIPItem("fe80::/10")},
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

func TestIntervalTreeContains(t *testing.T) {
	it := []tree.Interface{
		NewIPItem("0.0.0.0/4"),
		NewIPItem("127.0.0.0/8"),
		NewIPItem("192.168.0.0/16"),
		NewIPItem("192.169.0.0/16"),
		NewIPItem("192.170.0.0/16"),
		NewIPItem("192.171.0.0/16"),
		NewIPItem("::/4"),
		NewIPItem("fe80::/10"),
		NewIPItem("2001:db8::/32"),
	}

	tr, _ := tree.NewTree(it)

	valid := []struct {
		in   IPItem
		want IPItem
	}{
		{NewIPItem("0.0.0.0/6"), NewIPItem("0.0.0.0/4")},
		{NewIPItem("::/6"), NewIPItem("::/4")},
		{NewIPItem("127.0.0.1"), NewIPItem("127.0.0.0/8")},
		{NewIPItem("192.168.0.1"), NewIPItem("192.168.0.0/16")},
		{NewIPItem("192.169.0.0/16"), NewIPItem("192.169.0.0/16")},
		{NewIPItem("2001:db8::affe"), NewIPItem("2001:db8::/32")},
		{NewIPItem("::1"), NewIPItem("::/4")},
		{NewIPItem("fe80::/12"), NewIPItem("fe80::/10")},
	}

	for _, tt := range valid {
		got := tr.Superset(tt.in)
		if got != tt.want {
			t.Errorf("Contains(%v) = %v; want %v", tt.in, got, tt.want)
		}
	}

	invalid := []struct {
		in IPItem
	}{
		{NewIPItem("17.0.0.0/4")},
		{NewIPItem("::/3")},
		{NewIPItem("128.0.0.1")},
		{NewIPItem("168.192.0.1")},
		{NewIPItem("2001:db9::affe")},
		{NewIPItem("::/0")},
		{NewIPItem("fe80::/9")},
	}

	for _, tt := range invalid {
		got := tr.Superset(tt.in)
		if got != nil {
			t.Errorf("Contains(%v) =  %v; want: %v", tt.in, got, nil)
		}
	}

	if got := tr.Superset(nil); got != nil {
		t.Errorf("Contains(nil) = %v; want %v", got, nil)
	}
}

func TestIntervalTreeLookupMissing(t *testing.T) {
	it := []tree.Interface{
		NewIPItem("2001:db8::/32"),
		NewIPItem("::/0"),
		NewIPItem("::1"),
	}

	tr, _ := tree.NewTree(it)

	not := NewIPItem("1.2.3.4/5")
	if got := tr.Lookup(not); got != nil {
		t.Errorf("Lookup(%v) = %v; want: %v", not, got, nil)
	}
}

func TestIntervalTreeInsertOne(t *testing.T) {
	tr, _ := tree.NewTree([]tree.Interface{NewIPItem("0.0.0.0/0", "AllIPv4")})

	got := tr.String()

	want := `▼
└─ 0.0.0.0/0 > AllIPv4
`
	if got != want {
		t.Errorf("NewTree():\nwant: %v\ngot:  %v", want, got)
	}
}
