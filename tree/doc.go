/*
Package Tree is an implementation of a CIDR/Block prefix tree for fast IP lookup with longest-prefix-match.

It is NOT a standard patricia-trie, this isn't possible for general blocks not represented by bitmasks.

 type Tree struct {
  // Contains the root node the tree.
  Root *Node
 }

Node, recursive tree data structure. Items abstracted via Itemer interface

 type Node struct {
  Item   *Itemer
  Parent *Node
  Childs []*Node
 }

Itemer interface for Tree items, maybe with payload and not just ip Blocks.
See relation between Compare and Contains in the examples.
 type Itemer interface {

  // Contains, defines the depth in the tree, parent child relationship.
  Contains(Itemer) bool

  // Compare, defines equality and sort order on same tree level, siblings relationship.
  Compare(Itemer) int
 }

The tree can be visualized as:

 ▼
 ├─ 10.0.0.0/9
 │  ├─ 10.0.0.0/11
 │  │  ├─ 10.0.0.0/20
 │  │  ├─ 10.0.16.0/20
 │  │  └─ 10.0.32.0/20
 │  └─ 10.32.0.0/11
 │     ├─ 10.32.8.0/22
 │     ├─ 10.32.12.0/22
 │     └─ 10.32.16.0/22
 ├─ 2001:db8:900::/48
 │  ├─ 2001:db8:900::/49
 │  │  ├─ 2001:db8:900::/52

*/
package tree
