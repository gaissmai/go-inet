package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gaissmai/go-inet/inet"
	"github.com/gaissmai/go-inet/tree"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	bs := GenBlockItem(n)
	t := tree.NewBlockTree().InsertBulk(bs...)
	if n < 1000 {
		t.Fprint(os.Stdout)
	}

	for _, s := range []string{"134.60.2.83", "2001:7c0:900:4df::dfda"} {
		l := tree.NewSimpleItem(inet.MustIP(inet.NewIP(s)))
		if m, ok := t.Lookup(l); ok {
			fmt.Printf("%v found at %v\n", l, m)
		} else {
			fmt.Printf("%v not found\n", l)
		}
	}
}

// #####################################################################
// ### generators for IPs and CIDRs
// #####################################################################

func genV4(n int) []inet.IP {
	out := make([]inet.IP, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, rand.Uint32())
		ip, _ := inet.NewIP(buf)
		out[i] = ip
	}
	return out
}

func genV6(n int) []inet.IP {
	out := make([]inet.IP, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, 16)
		binary.BigEndian.PutUint64(buf[:8], rand.Uint64())
		binary.BigEndian.PutUint64(buf[8:], rand.Uint64())
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

// #####################################################################

func genBlockV4(n int) []inet.Block {
	set := make(map[inet.Block]bool, n)

	rs := make([]inet.Block, 0, n)
	for _, v := range genV4(n) {
		ones := rand.Intn(32)
		b, _ := inet.NewBlock(fmt.Sprintf("%s/%d", v, ones))
		if val := set[b]; val {
			continue
		}
		set[b] = true
		rs = append(rs, b)
	}
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	return rs
}

func genBlockV6(n int) []inet.Block {
	set := make(map[inet.Block]bool, n)

	rs := make([]inet.Block, 0, n)
	for _, v := range genV6(n) {
		ones := rand.Intn(128)
		b, _ := inet.NewBlock(fmt.Sprintf("%s/%d", v, ones))
		if val := set[b]; val {
			continue
		}
		set[b] = true
		rs = append(rs, b)
	}
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	return rs
}

func genBlockMixed(n int) []inet.Block {
	rs := make([]inet.Block, 0, n)
	rs = append(rs, genBlockV4(n/2)...)
	rs = append(rs, genBlockV6(n/2)...)
	rand.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
	return rs
}

// #####################################################################

func GenBlockItemV4(n int) []tree.Itemer {
	bs := genBlockV4(n)
	out := make([]tree.Itemer, 0, n)
	for _, b := range bs {
		out = append(out, tree.NewSimpleItem(b))
	}
	return out
}

func GenBlockItemV6(n int) []tree.Itemer {
	bs := genBlockV6(n)
	out := make([]tree.Itemer, 0, n)
	for _, b := range bs {
		out = append(out, tree.NewSimpleItem(b))
	}
	return out
}

// #############################################################################

func GenBlockItemIP(n int) []tree.Itemer {
	ips := genMixed(n)
	out := make([]tree.Itemer, 0, n)
	for _, ip := range ips {
		out = append(out, tree.NewSimpleItem(ip))
	}
	return out
}

func GenBlockItem(n int) []tree.Itemer {
	bs := genBlockMixed(n)
	out := make([]tree.Itemer, 0, n)
	for _, b := range bs {
		out = append(out, tree.NewSimpleItem(b))
	}
	return out
}
