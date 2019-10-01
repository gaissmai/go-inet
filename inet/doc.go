/*
Package inet represents IP-addresses and IP-Blocks as comparable types.
A tree implemetation for longest-prefix-match is included.

IP addresses and blocks can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.

IP addresses are represented as fixed arrays of 21 bytes, this ensures natural sorting (IPv4 < IPv6).

 type IP [21]byte

  IP[0]    = version information (4 or 6)
  IP[1:5]  = IPv4 address, if version == 4, else zero
  IP[5:21] = IPv6 address, if version == 6, else zero

Blocks are IP-networks or IP-ranges, e.g.

 192.168.0.1/24              // CIDR, network
 ::1/128                     // CIDR, network
 10.0.0.3-10.0.17.134        // range
 2001:db8::1-2001:db8::f6    // range

A Block is represented as a struct of three IP addresses:

 type Block struct {
  Base IP
  Last IP
  Mask IP  // may be zero for begin-end ranges
 }

Tree is an implementation of a CIDR/Block tree for fast IP lookup with longest-prefix-match.

 type Tree struct {
  // Contains the root node the tree.
  Root *Node
 }

Node, recursive tree data structure, only public for easy serialization, don't rely on it.
Items abstracted via Itemer interface

 type Node struct {
  Item   *Itemer
  Parent *Node
  Childs []*Node
 }

Itemer interface for Tree items, maybe with payload and not just ip Blocks.
See relation between Compare and Contains at inet.Block.Compare()
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

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

This is a package for system programming, all fields are public for easy and fast serialization without special treatment.

Anyway, you should not direct modify the fields and bytes, unless you know what you are doing.
*/
package inet
