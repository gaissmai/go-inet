package tree_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"testing"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

var r = rand.New(rand.NewSource(42))

func BenchmarkTreeInsert(b *testing.B) {
	bench := []int{1000, 10000, 100000, 1000000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		is := make([]tree.Item, len(bs))
		for i := range bs {
			is[i] = tree.Item{bs[i], nil, nil}
		}

		t := tree.New()

		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = t.Insert(is...)
			}
		})

	}
}

func BenchmarkLookupTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		is := make([]tree.Item, len(bs))
		for i := range bs {
			is[i] = tree.Item{bs[i], nil, nil}
		}

		t := tree.New()
		_ = t.Insert(is...)

		vx := is[rand.Intn(len(is))]
		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
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
		is := make([]tree.Item, len(bs))
		for i := range bs {
			is[i] = tree.Item{bs[i], nil, nil}
		}

		t := tree.New()
		_ = t.Insert(is...)

		var walkFn tree.WalkFunc = func(n *tree.Node, l int) error { return nil }

		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
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
		is := make([]tree.Item, len(bs))
		for i := range bs {
			is[i] = tree.Item{bs[i], nil, nil}
		}

		t := tree.New()
		_ = t.Insert(is...)

		vx := is[rand.Intn(len(is))]
		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.Remove(vx)
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
		ip, _ := inet.ParseIP(buf)
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
		ip, _ := inet.ParseIP(buf)
		out[i] = ip
	}
	return out
}

func genBlockV4(n int) []inet.Block {
	rs := make([]inet.Block, n)
	for i, v := range genV4(n) {
		ones := r.Intn(32)
		rs[i], _ = inet.ParseBlock(fmt.Sprintf("%s/%d", v, ones))
	}
	return rs
}

func genBlockV6(n int) []inet.Block {
	rs := make([]inet.Block, n)
	for i, v := range genV6(n) {
		ones := r.Intn(128)
		rs[i], _ = inet.ParseBlock(fmt.Sprintf("%s/%d", v, ones))
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
