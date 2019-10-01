## ATTENTION

The API is not stable yet!

# go-inet

A Go library for reading, formatting, sorting and converting IP-addresses and IP-blocks

This package represents IP-addresses and IP-Blocks as comparable types.
They can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.
A tree implemetation for longest-prefix-match is included.

## go-inet/inet

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
  Mask IP  // may be zero for begin-end ranges
 }
```

Tree is an implementation of a multi-root CIDR/Block tree for fast IP lookup with longest-prefix-match.

```go
 type Tree struct {
 	// Contains the root node of a multi-root tree.
 	// root-item and root-parent are nil for root-node.
 	Root *Node
 }
```

Node, recursive tree data structure, only public for easy serialization, don't rely on it.
Items are abstracted via Itemer interface

 ```go
 type Node struct {
 	Item   *Itemer
 	Parent *Node
 	Childs []*Node
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

These are packages for system programmers, all fields are public for easy and fast serialization without special treatment.
Anyway, you should not direct modify the fields and bytes, unless you know what you are doing.

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

