package inet

import (
	"math/rand"
	"net"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func mustBlock(i interface{}) Block {
	switch v := i.(type) {
	case string:
		if b, err := ParseBlock(v); err == nil {
			return b
		}
		panic(errInvalidBlock)
	case net.IPNet:
		if b, err := FromStdIPNet(v); err == nil {
			return b
		}
		panic(errInvalidBlock)
	default:
		panic(errInvalidBlock)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0.0.0.0/0", "0.0.0.0/0"},
		{"127.000.000.255/8", "127.0.0.0/8"},
		{"127.000.000.250-127.000.000.255", "127.0.0.250-127.0.0.255"},
		{"127.0.0.248-127.0.0.255", "127.0.0.248/29"},
		{"::ffff:182.239.134.2/32", "182.239.134.2/32"},
		{"2001:db8::/32", "2001:db8::/32"},
		{"::-::ffff", "::/112"},
	}

	for _, tt := range tests {
		r, err := ParseBlock(tt.in)
		if err != nil {
			t.Errorf("ParseBlock(%v) got error: %v", tt.in, err)
		}

		got := r.String()
		if got != tt.want {
			t.Errorf("Block(%q).String() != %v, got %v", tt.in, tt.want, got)
		}
	}
}

func TestParseFault(t *testing.T) {
	_, err := ParseBlock("")
	if err == nil {
		t.Errorf("ParseBlock(\"\"), expected error: %v", errInvalidBlock)
	}
}

func TestBaseLast(t *testing.T) {
	b, err := ParseBlock("1.2.3.4")
	if err != nil {
		t.Errorf("ParseBlock(IP) returns error: %v", err)
	}
	if b.Base() != b.Last() {
		t.Errorf("ParseBlock(IP), expected b.Base() == b.Last(), got (%v, %v)", b.Base(), b.Last())
	}
}

func TestFromStdlib(t *testing.T) {
	for _, tt := range []struct {
		in  net.IPNet
		err error
	}{
		{in: net.IPNet{IP: net.ParseIP("2001:db8::"), Mask: net.IPMask(net.ParseIP("fffe::"))}, err: nil},
		{in: net.IPNet{IP: net.ParseIP("1.2.3.400.500"), Mask: net.IPMask(net.ParseIP("255.0.0.0"))}, err: errInvalidBlock},
		{in: net.IPNet{IP: net.ParseIP("2001:db8::"), Mask: net.IPMask(net.ParseIP("Giraffe::"))}, err: errInvalidBlock},
	} {
		_, err := FromStdIPNet(tt.in)
		if err != tt.err {
			t.Errorf("Block from net.IPNet, got %v, want %v", err, tt.err)
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
		sortedBuf = append(sortedBuf, mustBlock(s))
	}

	// clone and shuffle
	mixedBuf := make([]Block, len(sortedBuf))
	copy(mixedBuf, sortedBuf)
	rand.Shuffle(len(mixedBuf), func(i, j int) { mixedBuf[i], mixedBuf[j] = mixedBuf[j], mixedBuf[i] })

	sort.Slice(mixedBuf, func(i, j int) bool { return mixedBuf[i].Less(mixedBuf[j]) })

	if !reflect.DeepEqual(mixedBuf, sortedBuf) {
		mixed := make([]string, 0, len(mixedBuf))
		for _, ipr := range mixedBuf {
			mixed = append(mixed, ipr.String())
		}

		t.Errorf("===> input:\n%v\n===> got:\n%v", strings.Join(sorted, "\n"), strings.Join(mixed, "\n"))
	}
}

func TestBlockCovers(t *testing.T) {
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

		ra := mustBlock(a)
		rb := mustBlock(b)

		got := rb.Covers(ra)
		if got != want {
			t.Errorf("(%v).Covers(%v) = %v; want %v", rb, ra, got, want)
		}
	}
}

func TestBlockMerge(t *testing.T) {
	bs := []Block{
		mustBlock("0.0.0.0/0"),
		mustBlock("10.0.0.15/32"),
		mustBlock("10.0.0.16/28"),
		mustBlock("10.0.0.32/27"),
		mustBlock("10.0.0.64/26"),
		mustBlock("10.0.0.128/26"),
		mustBlock("10.0.0.192/27"),
		mustBlock("134.60.0.0/16"),
		mustBlock("134.60.0.255/24"),
		mustBlock("193.197.62.192/29"),
		mustBlock("193.197.64.0/22"),
		mustBlock("193.197.228.0/22"),
		mustBlock("::/0"),
		mustBlock("::-::ffff"),
		mustBlock("2001:7c0:900::/48"),
		mustBlock("2001:7c0:900::/49"),
		mustBlock("2001:7c0:900::/52"),
		mustBlock("2001:7c0:900::/53"),
		mustBlock("2001:7c0:900:800::/56"),
		mustBlock("2001:7c0:900:800::/64"),
	}
	got := Merge(bs)

	want := []Block{
		mustBlock("0.0.0.0/0"),
		mustBlock("::/0"),
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge():\ngot:  %v\nwant: %v", got, want)
	}

	// corner cases
	bs = []Block{} // nil slice
	if got := Merge(bs); got != nil {
		t.Errorf("Merge() nil slice should return nil, got %v\n", got)
	}

	bs = []Block{mustBlock("0.0.0.0/8")}
	want = []Block{mustBlock("0.0.0.0/8")}
	got = Merge(bs)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Merge():\ngot:  %v\nwant: %v", got, want)
	}

}

func TestBlockToCIDRs(t *testing.T) {
	b, _ := ParseBlock("10.0.0.15-10.0.0.236")
	got := b.CIDRs()

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
		want = append(want, mustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.ToCIDRs(), got %v, want %v", b, got, want)
	}

	// corner case
	b = Block{}
	if b.CIDRs() != nil {
		t.Error("ToCIDRs on invalid block must reutn nil")
	}
}

func TestBlockToCIDRsV6(t *testing.T) {
	b, _ := ParseBlock("2001:db9::1-2001:db9::1234")
	got := b.CIDRs()

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
		want = append(want, mustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.ToCIDRs(), got %v, want %v", b, got, want)
	}
}

func TestBlockToCIDRsV4(t *testing.T) {
	b, _ := ParseBlock("255.255.255.253-255.255.255.255")
	got := b.CIDRs()

	var want []Block
	for _, s := range []string{
		"255.255.255.253/32",
		"255.255.255.254/31",
	} {
		want = append(want, mustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.ToCIDRs(), got %v, want %v", b, got, want)
	}
}

func TestBlockToCIDRsV6cornerCase(t *testing.T) {
	b, _ := ParseBlock("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffd-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")
	got := b.CIDRs()

	var want []Block
	for _, s := range []string{
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffd/128",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe/127",
	} {
		want = append(want, mustBlock(s))
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("%v.AsCIDRs(), got %v, want %v", b, got, want)
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

		ra := mustBlock(a)
		rb := mustBlock(b)

		got := ra.isDisjunct(rb)
		if got != want {
			t.Errorf("(%v).IsDisjunctWith(%v) = %v; want %v", ra, rb, got, want)
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

		ra := mustBlock(a)
		rb := mustBlock(b)

		got := ra.overlaps(rb)
		if got != want {
			t.Errorf("(%v).HasOverlapWith(%v) = %v; want %v", ra, rb, got, want)
		}
	}
}

// test the separation of the IPv4 and IPv6 address space
func TestBlockV4V6(t *testing.T) {
	r1 := mustBlock("0.0.0.0/0")
	r2 := mustBlock("::/0")

	if r1.overlaps(r2) != false {
		t.Errorf("%q.OverlapsWith(%q) == %t, want %t", r1, r2, r1.overlaps(r2), false)
	}
	if r2.overlaps(r1) != false {
		t.Errorf("%q.OverlapsWith(%q) == %t, want %t", r2, r1, r2.overlaps(r1), false)
	}
	if r2.Covers(r1) != false {
		t.Errorf("%q.Covers(%q) == %t, want %t", r2, r1, r2.Covers(r1), false)
	}
	if r1.Covers(r2) != false {
		t.Errorf("%q.Covers(%q) == %t, want %t", r1, r2, r1.Covers(r2), false)
	}
	if r1.isDisjunct(r2) != true {
		t.Errorf("%q.IsDisjunctWith(%q) == %t, want %t", r1, r2, r1.isDisjunct(r2), true)
	}
	if r2.isDisjunct(r1) != true {
		t.Errorf("%q.IsDisjunctWith(%q) == %t, want %t", r2, r1, r2.isDisjunct(r1), true)
	}
}

func TestFindDiffCornerCases(t *testing.T) {
	// nil
	r := mustBlock("::/0")
	rs := r.Diff(nil)

	if rs[0] != r {
		t.Errorf("Diff(nil), got %v, want %v", rs, []Block{r})
	}

	// self
	r = mustBlock("::/0")
	rs = r.Diff([]Block{r})
	if rs != nil {
		t.Errorf("Diff(self), got %v, want nil", rs)
	}

	// covers
	r = mustBlock("10.0.0.0/16")
	rs = r.Diff([]Block{mustBlock("10.0.0.0/8")})
	if rs != nil {
		t.Errorf("Diff(coverage), got %v, want nil", rs)
	}

	// overflow
	r = mustBlock("0.0.0.0/0")
	rs = r.Diff([]Block{mustBlock("255.255.255.255")})
	want := mustBlock("0.0.0.0-255.255.255.254")
	if rs[0] != want {
		t.Errorf("Diff(overflow), got %v, want %v", rs, want)
	}

	// base > last
	r = mustBlock("10.0.0.0/8")
	rs = r.Diff([]Block{mustBlock("10.128.0.0/9")})
	want = mustBlock("10.0.0.0/9")
	if rs[0] != want {
		t.Errorf("Diff(base>last), got %v, want %v", rs, want)
	}

}

func TestFindDiffIANAv6(t *testing.T) {
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
		inner = append(inner, mustBlock(s))
	}

	var want []Block
	for _, s := range []string{
		"6000::/3",
		"fc00::/7",
	} {
		want = append(want, mustBlock(s))
	}

	rs := b.Diff(inner)

	if !reflect.DeepEqual(rs, want) {
		t.Errorf("Diff for IANAv6 blocks, got %v, want %v", rs, want)
	}
}
