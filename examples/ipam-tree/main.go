// Command ipam-tree, read blocks from Stdin, print sorted Tree
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gaissmai/go-inet/v2/inettree"
	"github.com/gaissmai/go-inet/v2/tree"
)

func main() {
	r := csv.NewReader(os.Stdin)
	r.FieldsPerRecord = 2

	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	bs := make([]tree.Interface, 0, len(records))

	for i := range records {
		b := records[i][0]
		c := records[i][1]

		item, err := inettree.NewItem(b, c)
		if err != nil {
			log.Fatal(err)
		}
		text := item.Block.String()
		item.Text = text + " " + strings.Repeat(".", 40-len(text)) + " " + c
		bs = append(bs, item)
	}

	t, err := tree.NewTree(bs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(t)
}
