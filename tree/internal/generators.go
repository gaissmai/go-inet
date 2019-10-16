package internal

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/gaissmai/go-inet/inet"
)

var r = rand.New(rand.NewSource(42))

// #####################################################################
// ### generators for IPs and CIDRs
// #####################################################################

func GenV4(n int) []inet.IP {

	set := make(map[inet.IP]bool, n)
	for len(set) < n {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, r.Uint32())
		ip, _ := inet.ParseIP(buf)
		set[ip] = true
	}

	out := make([]inet.IP, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

func GenV6(n int) []inet.IP {
	set := make(map[inet.IP]bool, n)
	for len(set) < n {
		buf := make([]byte, 16)
		binary.BigEndian.PutUint64(buf[:8], r.Uint64())
		binary.BigEndian.PutUint64(buf[8:], r.Uint64())
		ip, _ := inet.ParseIP(buf)
		set[ip] = true
	}

	out := make([]inet.IP, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

func GenMixed(n int) []inet.IP {
	out := make([]inet.IP, 0, n)
	out = append(out, GenV4(n/2)...)
	out = append(out, GenV6(n/2)...)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func GenBlockV4(n int) []inet.Block {
	set := make(map[inet.Block]bool, n)

	for len(set) < n {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, r.Uint32())
		ip, _ := inet.ParseIP(buf)

		ones := r.Intn(24) + 8
		b, _ := inet.ParseBlock(fmt.Sprintf("%s/%d", ip, ones))

		set[b] = true
	}

	out := make([]inet.Block, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

func GenBlockV6(n int) []inet.Block {
	set := make(map[inet.Block]bool, n)

	for len(set) < n {
		buf := make([]byte, 16)
		binary.BigEndian.PutUint64(buf[:8], r.Uint64())
		binary.BigEndian.PutUint64(buf[8:], r.Uint64())
		ip, _ := inet.ParseIP(buf)

		ones := r.Intn(112) + 16
		b, _ := inet.ParseBlock(fmt.Sprintf("%s/%d", ip, ones))

		set[b] = true
	}

	out := make([]inet.Block, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

func GenBlockMixed(n int) []inet.Block {
	rs := make([]inet.Block, 0, n)
	rs = append(rs, GenBlockV4(n/2)...)
	rs = append(rs, GenBlockV6(n/2)...)
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

	return rs
}
