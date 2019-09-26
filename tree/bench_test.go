package tree

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/gaissmai/go-inet/inet"
)

func BenchmarkRemoveBlockTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		inet.SortBlock(bs)
		t := NewBlockTree()
		for i := range bs {
			t.Insert(bs[i])
		}

		vx := genBlockMixed(10)[0]
		b.Run(fmt.Sprintf("InsertRemoveBlockTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.Insert(vx)
				t.Remove(vx)
			}
		})

	}
}

func BenchmarkLookupBlockTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		inet.SortBlock(bs)
		t := NewBlockTree()
		for i := range bs {
			t.Insert(bs[i])
		}

		vx := bs[rand.Intn(len(bs))]
		b.Run(fmt.Sprintf("LookupBlockTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t.Lookup(vx)
			}
		})

	}
}

func BenchmarkWalkBlockTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		bs := genBlockMixed(n)
		inet.SortBlock(bs)
		t := NewBlockTree()
		for i := range bs {
			t.Insert(bs[i])
		}

		var walkFn WalkFunc = func(n *Node, l int) error { return nil }

		b.Run(fmt.Sprintf("WalkBlockTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = t.Walk(walkFn, false)
			}
		})

		b.Run(fmt.Sprintf("WalkSortedBlockTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = t.Walk(walkFn, true)
			}
		})

	}
}
