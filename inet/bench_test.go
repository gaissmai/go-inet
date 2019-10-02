package inet_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/gaissmai/go-inet/inet"
)

var r = rand.New(rand.NewSource(42))

func BenchmarkSortIP(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		ips := genMixed(n)
		rand.Shuffle(len(ips), func(i, j int) { ips[i], ips[j] = ips[j], ips[i] })

		b.Run(fmt.Sprintf("SortIP: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				inet.SortIP(ips)
			}
		})

	}
}

func BenchmarkSortBlock(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		rs := genBlockMixed(n)
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

		b.Run(fmt.Sprintf("SortBlock: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				inet.SortBlock(rs)
			}
		})

	}
}

func BenchmarkFindFreeCIDRv4(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		rs := genBlockV4(n)
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

		r, _ := inet.NewBlock("0.0.0.0/0")
		b.Run(fmt.Sprintf("FindFreeCIDRv4: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r.FindFreeCIDR(rs)
			}
		})

	}
}

func BenchmarkFindFreeCIDRv6(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		rs := genBlockV6(n)
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

		r, _ := inet.NewBlock("::/0")
		b.Run(fmt.Sprintf("FindFreeCIDRv6: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r.FindFreeCIDR(rs)
			}
		})

	}
}

// #####################################################################
// ### generators for IPs and CIDRs
// #####################################################################

func genV4(n int) []inet.IP {
	out := make([]inet.IP, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, r.Uint32())
		ip, _ := inet.NewIP(buf)
		out[i] = ip
	}
	return out
}

func genV6(n int) []inet.IP {
	out := make([]inet.IP, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 16)
		binary.BigEndian.PutUint64(buf[:8], r.Uint64())
		binary.BigEndian.PutUint64(buf[8:], r.Uint64())
		ip, _ := inet.NewIP(buf)
		out[i] = ip
	}
	return out
}

func genMixed(n int) []inet.IP {
	out := make([]inet.IP, 0, n)
	out = append(out, genV4(n/2)...)
	out = append(out, genV6(n/2)...)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func genBlockV4(n int) []inet.Block {
	rs := make([]inet.Block, n)
	for i, v := range genV4(n) {
		ones := r.Intn(32)
		rs[i], _ = inet.NewBlock(fmt.Sprintf("%s/%d", v, ones))
	}
	return rs
}

func genBlockV6(n int) []inet.Block {
	rs := make([]inet.Block, n)
	for i, v := range genV6(n) {
		ones := r.Intn(128)
		rs[i], _ = inet.NewBlock(fmt.Sprintf("%s/%d", v, ones))
	}
	return rs
}

func genBlockMixed(n int) []inet.Block {
	rs := make([]inet.Block, 0, n)
	rs = append(rs, genBlockV4(n/2)...)
	rs = append(rs, genBlockV6(n/2)...)
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	return rs
}

func BenchmarkLookupTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		t := inet.NewTree().InsertBulk(bs)

		vx := bs[rand.Intn(len(bs))]
		b.Run(fmt.Sprintf("LookupTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.Lookup(vx)
			}
		})

	}
}

func BenchmarkWalkTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		t := inet.NewTree().InsertBulk(bs)

		var walkFn inet.WalkFunc = func(n *inet.Node, l int) error { return nil }

		b.Run(fmt.Sprintf("WalkTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = t.Walk(walkFn)
			}
		})

	}
}

func BenchmarkTreeRemoveItem(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		t := inet.NewTree().InsertBulk(bs)

		vx := bs[rand.Intn(len(bs))]
		b.Run(fmt.Sprintf("RemoveItem: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.Remove(vx)
			}
		})

	}
}
