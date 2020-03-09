package inet

import (
	"math/rand"
	"net"
	"reflect"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"127.000.000.255/8", "127.0.0.0/8"},
		{"127.000.000.250-127.000.000.255", "127.0.0.250-127.0.0.255"},
		{"127.0.0.248-127.0.0.255", "127.0.0.248/29"},
		{"::ffff:182.239.134.2/32", "182.239.134.2/32"},
	}

	for _, tt := range tests {
		r, _ := ParseBlock(tt.in)
		got := r.String()
		if got != tt.want {
			t.Errorf("Block(%q).String() != %v, got %v", tt.in, tt.want, got)
		}
	}
}

func TestFromStdlib(t *testing.T) {
	tests := []interface{}{
		net.IP([]byte{10, 0, 0, 1}),
		net.IP([]byte{0x20, 0x01, 0x0d, 0xdb}),
		net.IPNet{
			IP:   net.IP([]byte{0: 0x20, 1: 0x01, 2: 0x0d, 3: 0xdb, 15: 0}),
			Mask: net.IPMask([]byte{0: 0xff, 1: 0xfe, 15: 0}),
		},
	}

	for _, ip := range tests {
		_, err := ParseBlock(ip)
		if err != nil {
			t.Errorf("Block from net.IP, got error %s", err)
		}
	}
}

func TestParseBlockFail(t *testing.T) {
	tests := []string{
		"-10.0.0.0",
		"/32",
		"10.0.0.0/33",
		"127.355.0.1/8",
		"127.0.0.3-127.0.0.2",
		"315.0.0.3-127.0.0.2",
		"127.0.0.3-127.0.0.256",
		"2001:gb8::/32",
		"2001:db8::-2001:db7::ffff",
		"2001:dx8::-2001:db8::",
		"2001:db8::-2001:db8::x",
		"127.0.0.1-2001:db8::",
		"127.0.0.1::127.0.0.17",
	}

	for _, in := range tests {
		_, err := ParseBlock(in)

		if err == nil {
			t.Errorf("success for ParseBlock(%s) is not expected!", in)
		}
	}
}

func TestBlockIsValid(t *testing.T) {
	r1 := MustBlock("127.0.0.1/8")
	r2 := MustBlock("127.0.0.0-127.255.255.254")
	r3 := MustBlock("2001:db8::/127")
	r4 := MustBlock("2001:db8::-2001:dff::")

	for _, b := range []Block{r1, r2, r3, r4} {
		if !b.IsValid() {
			t.Errorf("b.IsValid returns false, want true")
		}
	}

	r1.Base = r1.Last.AddUint64(1)
	r2.Last = IP{}
	r3.Mask = r3.Mask.SubUint64(1)
	r4.Mask[2] = 0xff

	for _, b := range []Block{r1, r2, r3, r4} {
		if b.IsValid() {
			t.Errorf("b.IsValid returns true, want false")
		}
	}

}

func TestBlockHasOverlapWith(t *testing.T) {

	tests := []struct {
		a, b string
		want bool
	}{
		{
			a:    "::/0", // v4 v6 mismatch
			b:    "0.0.0.0/0",
			want: false,
		},
		{
			a:    "127.0.0.0/8",
			b:    "10.0.0.0/8",
			want: false,
		},
		{
			a:    "10.0.0.0/8",
			b:    "127.0.0.0/8",
			want: false,
		},
		{
			a:    "127.0.0.0-127.0.0.255",
			b:    "127.0.0.128-128.0.0.100",
			want: true,
		},
		{
			a:    "127.0.0.128-128.0.0.100",
			b:    "127.0.0.0-127.0.0.255",
			want: true,
		},
		{
			a:    "10.0.0.0/30",
			b:    "10.0.0.3-10.0.0.14",
			want: true,
		},
		{
			a:    "10.0.0.0-10.0.0.3",
			b:    "10.0.0.3-10.0.0.14",
			want: true,
		},
		{
			a:    "::/0",
			b:    "::/0",
			want: false,
		},
		{
			a:    "::/0",
			b:    "::/60",
			want: false,
		},
		{
			a:    "4.4.4.4/32",
			b:    "4.4.4.4/32",
			want: false,
		},
	}

	for _, tt := range tests {
		a, b, want := tt.a, tt.b, tt.want

		ra := MustBlock(a)
		rb := MustBlock(b)

		got := ra.OverlapsWith(rb)
		if got != want {
			t.Errorf("(%v).HasOverlapWith(%v) = %v; want %v", ra, rb, got, want)
		}
	}
}

func TestBlockIsDisjunctWith(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{
			a:    "0.0.0.0/0",
			b:    "::/0",
			want: true,
		},
		{
			a:    "10.0.0.0/8",
			b:    "fe80::/12",
			want: true,
		},
		{
			a:    "0.0.0.0/0",
			b:    "10.0.0.0/8",
			want: false,
		},
		{
			a:    "10.0.0.0/8",
			b:    "10.0.0.0/7",
			want: false,
		},
		{
			a:    "10.0.0.0/7",
			b:    "10.0.0.0/8",
			want: false,
		},
		{
			a:    "10.0.0.0/31",
			b:    "10.0.0.2/31",
			want: true,
		},
		{
			a:    "10.0.0.0-10.0.0.1",
			b:    "10.0.0.2/31",
			want: true,
		},
		{
			a:    "2001:db8::/127",
			b:    "2001:db8::2/127",
			want: true,
		},
		{
			a:    "10.0.0.3-10.0.0.14",
			b:    "10.0.0.0/30",
			want: false,
		},
		{
			a:    "10.0.0.3-10.0.0.14",
			b:    "10.0.0.0-10.0.0.3",
			want: false,
		},
	}

	for _, tt := range tests {
		a, b, want := tt.a, tt.b, tt.want

		ra := MustBlock(a)
		rb := MustBlock(b)

		got := ra.IsDisjunctWith(rb)
		if got != want {
			t.Errorf("(%v).IsDisjunctWith(%v) = %v; want %v", ra, rb, got, want)
		}
	}
}

func TestBlockContains(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{
			a:    "::/0",
			b:    "0.0.0.0/0",
			want: false,
		},
		{
			a:    "0.0.0.0/0",
			b:    "0.0.0.0/0",
			want: false,
		},
		{
			a:    "::/0",
			b:    "::/0",
			want: false,
		},
		{
			a:    "10.0.0.0/8",
			b:    "0.0.0.0/0",
			want: true,
		},
		{
			a:    "0.0.0.0/0",
			b:    "10.0.0.0/8",
			want: false,
		},
		{
			a:    "fe80::/12",
			b:    "::/0",
			want: true,
		},
		{
			a:    "::/0",
			b:    "fe80::/12",
			want: false,
		},
		{
			a:    "fe81::/16",
			b:    "fe80::/16",
			want: false,
		},
	}

	for _, tt := range tests {
		a, b, want := tt.a, tt.b, tt.want

		ra := MustBlock(a)
		rb := MustBlock(b)

		got := rb.Contains(ra)
		if got != want {
			t.Errorf("(%v).Contains(%v) = %v; want %v", rb, ra, got, want)
		}
	}
}

func TestSortBlock(t *testing.T) {

	sorted := []string{
		"0.0.0.0/0",
		"10.0.0.0/9",
		"10.0.0.0/11",
		"10.32.0.0/11",
		"10.64.0.0/11",
		"10.96.0.0/11",
		"10.96.0.0/13",
		"10.96.0.2-10.96.1.17",
		"10.104.0.0/13",
		"10.112.0.0/13",
		"10.120.0.0/13",
		"10.128.0.0/9",
		"134.60.0.0/16",
		"193.197.62.192/29",
		"193.197.64.0/22",
		"193.197.228.0/22",
		"::/0",
		"::-::ffff",
		"2001:7c0:900::/48",
		"2001:7c0:900::/49",
		"2001:7c0:900::/52",
		"2001:7c0:900::/53",
		"2001:7c0:900:800::/56",
		"2001:7c0:900:800::/64",
		"2001:7c0:900:1000::/52",
		"2001:7c0:900:1000::/53",
		"2001:7c0:900:1800::/53",
		"2001:7c0:900:8000::/49",
		"2001:7c0:900:8000::/56",
		"2001:7c0:900:8100::/56",
		"2001:7c0:2330::/44",
	}

	var sortedBuf []Block
	for _, s := range sorted {
		sortedBuf = append(sortedBuf, MustBlock(s))
	}

	// clone and shuffle
	mixedBuf := make([]Block, len(sortedBuf))
	copy(mixedBuf, sortedBuf)
	rand.Shuffle(len(mixedBuf), func(i, j int) { mixedBuf[i], mixedBuf[j] = mixedBuf[j], mixedBuf[i] })

	SortBlock(mixedBuf)

	if !reflect.DeepEqual(mixedBuf, sortedBuf) {
		mixed := make([]string, 0, len(mixedBuf))
		for _, ipr := range mixedBuf {
			mixed = append(mixed, ipr.String())
		}

		t.Errorf("===> input:\n%v\n===> got:\n%v", strings.Join(sorted, "\n"), strings.Join(mixed, "\n"))
	}
}

func TestSplitBlockZero(t *testing.T) {
	b := Block{}
	splits := b.SplitCIDR(1)

	if splits != nil {
		t.Errorf("error in splitting blockZero, got: %v, want nil)", splits)
	}
}

func TestSplitMaskZero(t *testing.T) {
	var r Block
	r.Base[0] = 4
	r.Last[0] = 4

	// Mask is still blockZero, we can't cplit without a mask
	splits := r.SplitCIDR(1)

	if splits != nil {
		t.Errorf("error in splitting a non CIDR range, got: %v, want nil)", splits)
	}
}

func TestBlockMarshalText(t *testing.T) {
	// test failure modes

	b := Block{}
	bs, _ := b.MarshalText()
	if len(bs) != 0 {
		t.Errorf("MarshalText for zero-value must return an empty []byte, got %#v", bs)
	}
}

// test the separation of the IPv4 and IPv6 address space
func TestBlockV4V6(t *testing.T) {
	r1 := MustBlock("0.0.0.0/0")
	r2 := MustBlock("::/0")

	if r1.OverlapsWith(r2) != false {
		t.Errorf("%q.OverlapsWith(%q) == %t, want %t", r1, r2, r1.OverlapsWith(r2), false)
	}
	if r2.OverlapsWith(r1) != false {
		t.Errorf("%q.OverlapsWith(%q) == %t, want %t", r2, r1, r2.OverlapsWith(r1), false)
	}
	if r2.Contains(r1) != false {
		t.Errorf("%q.Contains(%q) == %t, want %t", r2, r1, r2.Contains(r1), false)
	}
	if r1.Contains(r2) != false {
		t.Errorf("%q.Contains(%q) == %t, want %t", r1, r2, r1.Contains(r2), false)
	}
	if r1.IsDisjunctWith(r2) != true {
		t.Errorf("%q.IsDisjunctWith(%q) == %t, want %t", r1, r2, r1.IsDisjunctWith(r2), true)
	}
	if r2.IsDisjunctWith(r1) != true {
		t.Errorf("%q.IsDisjunctWith(%q) == %t, want %t", r2, r1, r2.IsDisjunctWith(r1), true)
	}
}

func TestFindFreeCIDRBlockNil(t *testing.T) {
	r := MustBlock("::/0")
	rs := r.FindFreeCIDR(nil)

	if rs[0] != r {
		t.Errorf("FindFreeCIDR for inner == nil, got %v, want %v", rs, []Block{r})
	}
}

func TestFindFreeCIDRBlockSelf(t *testing.T) {
	r := MustBlock("::/0")
	rs := r.FindFreeCIDR([]Block{r})
	if rs != nil {
		t.Errorf("FindFreeCIDR for inner == self, got %#v, want nil", rs)
	}
}

func TestFindFreeCIDRBlockIANAv6(t *testing.T) {
	b, _ := ParseBlock("::/0")

	var inner []Block
	for _, s := range []string{
		"0000::/8",
		"0100::/8",
		"0200::/7",
		"0400::/6",
		"0800::/5",
		"1000::/4",
		"2000::/3",
		"4000::/3",
		// "6000::/3",
		"8000::/3",
		"a000::/3",
		"c000::/3",
		"e000::/4",
		"f000::/5",
		"f800::/6",
		// "fc00::/7",
		"fe00::/9",
		"fe80::/10",
		"fec0::/10",
		"ff00::/8",
	} {
		inner = append(inner, MustBlock(s))
	}

	var want []Block
	for _, s := range []string{
		"6000::/3",
		"fc00::/7",
	} {
		want = append(want, MustBlock(s))
	}

	rs := b.FindFreeCIDR(inner)

	if !reflect.DeepEqual(rs, want) {
		t.Errorf("FindFreeCIDR for IANAv6 blocks, got %v, want %v", rs, want)
	}
}

func TestBlockToCIDRListV4(t *testing.T) {
	b, _ := ParseBlock("10.0.0.15-10.0.0.236")
	got := b.BlockToCIDRList()

	var want []Block
	for _, s := range []string{
		"10.0.0.15/32",
		"10.0.0.16/28",
		"10.0.0.32/27",
		"10.0.0.64/26",
		"10.0.0.128/26",
		"10.0.0.192/27",
		"10.0.0.224/29",
		"10.0.0.232/30",
		"10.0.0.236/32",
	} {
		want = append(want, MustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.BlockToCIDRList(), got %v, want %v", b, got, want)
	}
}

func TestBlockToCIDRListV6(t *testing.T) {
	b, _ := ParseBlock("2001:db9::1-2001:db9::1234")
	got := b.BlockToCIDRList()

	var want []Block
	for _, s := range []string{
		"2001:db9::1/128",
		"2001:db9::2/127",
		"2001:db9::4/126",
		"2001:db9::8/125",
		"2001:db9::10/124",
		"2001:db9::20/123",
		"2001:db9::40/122",
		"2001:db9::80/121",
		"2001:db9::100/120",
		"2001:db9::200/119",
		"2001:db9::400/118",
		"2001:db9::800/117",
		"2001:db9::1000/119",
		"2001:db9::1200/123",
		"2001:db9::1220/124",
		"2001:db9::1230/126",
		"2001:db9::1234/128",
	} {
		want = append(want, MustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.BlockToCIDRList(), got %v, want %v", b, got, want)
	}
}

func TestBlockToCIDRListV4Overflow(t *testing.T) {
	b, _ := ParseBlock("255.255.255.253-255.255.255.255")
	got := b.BlockToCIDRList()

	var want []Block
	for _, s := range []string{
		"255.255.255.253/32",
		"255.255.255.254/31",
	} {
		want = append(want, MustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.BlockToCIDRList(), got %v, want %v", b, got, want)
	}
}

func TestBlockToCIDRListV6Overflow(t *testing.T) {
	b, _ := ParseBlock("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffd-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	got := b.BlockToCIDRList()

	var want []Block
	for _, s := range []string{
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffd/128",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe/127",
	} {
		want = append(want, MustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.BlockToCIDRList(), got %v, want %v", b, got, want)
	}
}
