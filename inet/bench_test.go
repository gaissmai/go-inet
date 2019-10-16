package inet_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/inet/internal"
)

func BenchmarkSortIP(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		ips := internal.GenMixed(n)
		rand.Shuffle(len(ips), func(i, j int) { ips[i], ips[j] = ips[j], ips[i] })

		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				inet.SortIP(ips)
			}
		})

	}
}

func BenchmarkSortBlock(b *testing.B) {
	bench := []int{10000, 100000, 1000000}

	for _, n := range bench {
		rs := internal.GenBlockMixed(n)
		rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

		b.Run(fmt.Sprintf("%7d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				inet.SortBlock(rs)
			}
		})

	}
}
