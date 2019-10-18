package tree

import (
	"fmt"
	"sort"
)

func (t *Tree) Remove2(item Item) bool {
	return t.Root.remove2(item, false)
}

// recursive work horse
func (node *Node) remove2(input Item, removeBranch bool) bool {

	// childs are sorted, find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Block) >= 0 })

	if idx != l && input.Block.Compare(node.Childs[idx].Item.Block) == 0 {
		// found child to remove at [idx]
		// delete this child [idx] from node
		child := node.Childs[idx]

		if idx < l-1 {
			node.Childs = append(node.Childs[:idx], node.Childs[idx+1:]...)
		} else {
			node.Childs = node.Childs[:idx]
		}

		// remove branch, stop here
		if removeBranch {
			return true
		}

		// re-insert grandChilds from removed child into tree
		for _, grandChild := range child.Childs {

			// insert this grandchild at outer node
			if err := node.insertNode(grandChild); err != nil {
				panic(err)
			}
		}

		return true
	}

	// pos in tree not found on this level
	// walk down if any child (before input, respect sort order) includes input
	for j := idx - 1; j >= 0; j-- {
		child := node.Childs[j]
		if child.Item.Block.Contains(input.Block) {
			return child.remove(input, removeBranch)
		}
	}

	// not equal to any child and not contained in any child
	return false
}

func (node *Node) insertNode(input *Node) error {

	// childs are sorted find pos in childs on this level, binary search
	l := len(node.Childs)
	idx := sort.Search(l, func(i int) bool { return node.Childs[i].Item.Block.Compare(input.Item.Block) >= 0 })

	// not at end of slice
	if idx < l {
		// don't insert dups
		if input.Item.Block.Compare(node.Childs[idx].Item.Block) == 0 {
			return fmt.Errorf("duplicate item: %s", input.Item)
		}
	}

	// not in front of slice, check if previous child contains this new node
	if idx > 0 {
		child := node.Childs[idx-1]
		if child.Item.Block.Contains(input.Item.Block) {
			return child.insertNode(input)
		}
	}

	// add as new child on this level
	input.Parent = node

	// input is greater than all others and not contained, just append
	if idx == l {
		node.Childs = append(node.Childs, input)
		return nil
	}

	// buffer to build resorted childs
	buf := make([]*Node, 0, l+1)

	// copy [:idx] to buf
	buf = append(buf, node.Childs[:idx]...)

	// copy x to buf at [idx]
	buf = append(buf, input)

	// now handle [idx:]
	// resort if input contains next child...
	j := idx
	for {
		child := node.Childs[j]
		if input.Item.Block.Contains(child.Item.Block) {
			// recursive call, put next child in row under new input.Item
			if err := input.insertNode(child); err != nil {
				panic(err)
			}
			if j++; j < l {
				continue
			}
		}

		// childs are sorted, break after first child not being child of input
		break
	}

	// copy rest of childs to buf
	buf = append(buf, node.Childs[j:]...)

	node.Childs = buf
	return nil
}

