package tree

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkInsertBulkBlockTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		blocks := GenSimpleItem(n)

		b.Run(fmt.Sprintf("InsertBulkBlockTree: %d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				t := NewBlockTree()
				t.InsertBulk(blocks...)
			}
		})
	}
}

func BenchmarkRemoveBlockTree(b *testing.B) {
	bench := []int{1000, 10000, 100000}

	for _, n := range bench {
		blocks := GenSimpleItem(n)

		t := NewBlockTree()
		t.InsertBulk(blocks...)

		vx := GenSimpleItem(10)[0]
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
		blocks := GenSimpleItem(n)

		t := NewBlockTree()
		t.InsertBulk(blocks...)

		vx := blocks[rand.Intn(len(blocks))]
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
		blocks := GenSimpleItem(n)

		t := NewBlockTree()
		t.InsertBulk(blocks...)

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
