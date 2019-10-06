## RELASE NOTES

The API is now almost stable.

# go-inet

A Go library for reading, formatting, sorting and converting IP-addresses and IP-blocks.
A Tree implementation for lookups with longest-prefix-match is included.

## go-inet/inet

Package inet represents IP-addresses and IP-Blocks as comparable types.
A tree implementation for longest-prefix-match is included.

IP addresses and blocks can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.
A Tree implementation for lookups with longest-prefix-match is included.

This is a package for system programming, all fields are public for easy and fast serialization without special treatment. Anyway, you should not direct modify the fields and bytes, unless you know what you are doing.

IP addresses are represented as fixed arrays of 21 bytes, this ensures natural sorting (IPv4 < IPv6).

```go
  type IP [21]byte

  // IP[0]    = version information (4 or 6)
  // IP[1:5]  = IPv4 address, if version == 4, else zero
  // IP[5:21] = IPv6 address, if version == 6, else zero
```

Blocks are IP-networks or IP-ranges, e.g.

    192.168.0.1/24              // CIDR, network
    ::1/128                     // CIDR, network
    10.0.0.3-10.0.17.134        // range
    2001:db8::1-2001:db8::f6    // range

Blocks are represented as a struct of three IP addresses:

```go
  type Block struct {
    Base IP
    Last IP
    Mask IP // may be zero for begin-end ranges
  }
```

Tree is an implementation of a CIDR/Block tree for fast IP lookup with longest-prefix-match.
It is *NOT* a radix-tree, not possible for general IP blocks not represented by bitmasks.

```go
  type Tree struct {
    Root *Node
  }
```

Node, recursive data structure. Items are abstracted via Itemer interface

```go
  type Node struct {
    Item   *Itemer
    Parent *Node
    Childs []*Node
  }
```

Itemer interface for Tree items, maybe with payload and not just ip Blocks.

```go
  type Itemer interface {
   
   // Contains, defines the depth in the tree, parent child relationship.
   Contains(Itemer) bool
   
   // Compare, defines equality and sort order on same tree level, siblings relationship.
   Compare(Itemer) int
  }
```

The tree can be visualized as:

```
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
```

## Documentation

[![GoDoc](https://godoc.org/github.com/gaissmai/go-inet?status.svg)](https://godoc.org/github.com/gaissmai/go-inet)

Full `go doc` style documentation for the project can be viewed online without
installing this package by using the excellent GoDoc site here:
http://godoc.org/github.com/gaissmai/go-inet


## Installation

```bash
$ go get -u github.com/gaissmai/go-inet/...
```
You can also view the documentation locally once the package is installed with
the `godoc` tool by running

```bash
$ godoc -http=:6060
```
and pointing your browser to
http://localhost:6060/pkg/github.com/gaissmai/go-inet

## License

go-inet is licensed under the MIT License.

