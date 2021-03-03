// Package internal for generating random ip addresses and CIDRs for benchmarking.
package internal

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/gaissmai/go-inet/inet"
)

// #####################################################################
// ### generators for IPs and CIDRs
// #####################################################################

// GenV4 returns random v4 IP addresses.
func GenV4(n int) []inet.IP {

	var r = rand.New(rand.NewSource(int64(n)))

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

// GenV6 returns random v6 IP addresses.
func GenV6(n int) []inet.IP {
	var r = rand.New(rand.NewSource(int64(n)))

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

// GenMixed returns random mixed v4/v6 IP addresses.
func GenMixed(n int) []inet.IP {
	out := make([]inet.IP, 0, n)
	out = append(out, GenV4(n/2)...)
	out = append(out, GenV6(n/2)...)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// GenBlockV4 returns random v4 CIDRs
func GenBlockV4(n int) []inet.Block {
	var r = rand.New(rand.NewSource(int64(n)))
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

// GenBlockV6 returns random v6 CIDRs
func GenBlockV6(n int) []inet.Block {
	var r = rand.New(rand.NewSource(int64(n)))
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

// GenBlockMixed returns random mixed v4/v6 CIDRs
func GenBlockMixed(n int) []inet.Block {
	rs := make([]inet.Block, 0, n)
	rs = append(rs, GenBlockV4(n/2)...)
	rs = append(rs, GenBlockV6(n/2)...)
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

	return rs
}

// GenRangeV4 returns random v4 ranges
func GenRangeV4(n int) []inet.Block {
	var r = rand.New(rand.NewSource(int64(n)))
	set := make(map[inet.Block]bool, n)

	for len(set) < n {
		buf1 := make([]byte, 4)
		buf2 := make([]byte, 4)
		binary.BigEndian.PutUint32(buf1, r.Uint32())
		binary.BigEndian.PutUint32(buf2, r.Uint32())
		ip1, _ := inet.ParseIP(buf1)
		ip2, _ := inet.ParseIP(buf2)
		if ip1 > ip2 {
			ip1, ip2 = ip2, ip1
		}

		b, _ := inet.ParseBlock(fmt.Sprintf("%s-%s", ip1, ip2))
		set[b] = true
	}

	out := make([]inet.Block, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

// GenRangeV6 returns random v6 ranges
func GenRangeV6(n int) []inet.Block {
	var r = rand.New(rand.NewSource(int64(n)))
	set := make(map[inet.Block]bool, n)

	for len(set) < n {
		buf1 := make([]byte, 16)
		buf2 := make([]byte, 16)

		binary.BigEndian.PutUint64(buf1[:8], r.Uint64())
		binary.BigEndian.PutUint64(buf1[8:], r.Uint64())

		binary.BigEndian.PutUint64(buf2[:8], r.Uint64())
		binary.BigEndian.PutUint64(buf2[8:], r.Uint64())

		ip1, _ := inet.ParseIP(buf1)
		ip2, _ := inet.ParseIP(buf2)

		if ip1 > ip2 {
			ip1, ip2 = ip2, ip1
		}

		b, _ := inet.ParseBlock(fmt.Sprintf("%s-%s", ip1, ip2))
		set[b] = true
	}

	out := make([]inet.Block, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	return out
}

// GenRangeMixed returns random mixed v4/v6 ranges
func GenRangeMixed(n int) []inet.Block {
	rs := make([]inet.Block, 0, n)
	rs = append(rs, GenRangeV4(n/2)...)
	rs = append(rs, GenRangeV6(n/2)...)
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })

	return rs
}
