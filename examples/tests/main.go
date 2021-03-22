package main

import (
	"log"
	"os"

	"github.com/gaissmai/go-inet/v2/inet"
	// "github.com/gaissmai/go-inet/v2/inettree"
	// "github.com/gaissmai/go-inet/v2/tree"
)

func main() {
	s := os.Args[1]
	b, err := inet.ParseBlock(s)
	log.Printf("error: %v", err)
	log.Printf("b:     %v", b)

}
