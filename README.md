# go-inet
A Go library for reading, formatting, sorting and converting IP-addresses and IP-blocks
=======
## ATTENTION

The API is **NOT** **stable** yet!

# go-inet/inet

A Go library for reading, formatting, sorting and converting IP-addresses and IP-blocks.

This package represents IP-addresses and IP-Blocks as comparable types.
They can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.

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
Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

# go-inet/tree

An IP CIDR/Block-Tree is supported for fast and easy IP address lookups similar to routing tables.

The tree can be visualized as:

```
▼
├─ ::/8.................   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 100::/8..............   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 200::/7..............   "Reserved by IETF     [RFC4048]"
├─ 400::/6..............   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 800::/5..............   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 1000::/4.............   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 2000::/3.............   "Global Unicast       [RFC3513][RFC4291]"
│  ├─ 2000::/4.............  "Test"
│  └─ 3000::/4.............  "FREE"
├─ 4000::/3.............   "Reserved by IETF     [RFC3513][RFC4291]"
├─ 6000::/3.............   "Reserved by IETF     [RFC3513][RFC4291]"
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

