# go-inet
[![Build Status](https://img.shields.io/travis/gaissmai/go-inet.svg)](https://travis-ci.org/gaissmai/go-inet)
[![Go Reference](https://pkg.go.dev/badge/github.com/gaissmai/go-inet.svg)](https://pkg.go.dev/github.com/gaissmai/go-inet/v2)
[![GitHub](https://img.shields.io/github/license/gaissmai/go-inet)](https://github.com/gaissmai/go-inet/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaissmai/go-inet)](https://goreportcard.com/report/github.com/gaissmai/go-inet/v2)
[![Coverage Status](https://coveralls.io/repos/github/gaissmai/go-inet/badge.svg)](https://coveralls.io/github/gaissmai/go-inet/)

## ATTENTION: v2 is coming ...

A Go library for reading, formatting, sorting and converting IP-addresses and IP-blocks.

Version v2 uses the math based on `type uint128 struct {hi uint64, lo uint64}`, no longer bytes fiddling in network byte order.
The API is reduced to the bare minimum, the tree representation is abstracted with an Interface.

## go-inet/inet

Package inet represents IP-addresses and IP-Blocks as comparable types.

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

This IP representation is comparable and can be sorted very quickly without prior conversions to/from the different IP versions.

The library is mainly intended for fast ACL-lookups and for IP address management (IPAM) in global scope
and not for host related systems programming.

So, no IP address zone indices are supported and IPv4-mapped IPv6 addresses are stripped down to plain IPv4 addresses.
The information of the prior mapping is discarded.

Blocks are IP-networks or arbitrary IP-ranges, e.g.

    192.168.0.1/24              // CIDR
    ::1/128                     // CIDR
    10.0.0.3-10.0.17.134        // IP range, no CIDR
    2001:db8::1-2001:db8::f6    // IP range, no CIDR

## go-inet/tree

Package tree is a minimal interval tree implementation.

All interval types implementing the tree.Interface can use this library for fast lookups
and a stringified tree representation.

The tree can be visualized as:

```
 ▼
 ├─ 10.0.0.0/9
 │  ├─ 10.0.0.0/11
 │  │  ├─ 10.0.0.0-10.0.0.29
 │  │  ├─ 10.0.16.0/20
 │  │  └─ 10.0.32.0/20
 │  └─ 10.32.0.0/11
 │     ├─ 10.32.8.0/22
 │     ├─ 10.32.12.0-10.32.13.77
 │     └─ 10.32.16.0/22
 ├─ 2001:db8:900::/48
 │  ├─ 2001:db8:900::/49
 │  │  ├─ 2001:db8:900::/52
```

## go-inet/inettree

Package inettree implements the tree.Interface for inet.Block

## Documentation


Full `go doc` style documentation for the project can be viewed online without
installing this package by using the excellent [Pkg.go.dev site](https://pkg.go.dev/github.com/gaissmai/go-inet/v2)

## Installation

```bash
$ go get -u github.com/gaissmai/go-inet/v2/...
```
You can also view the documentation locally once the package is installed with
the `godoc` tool by running

```bash
$ godoc -http=:6060
```
and pointing your browser to
http://localhost:6060/pkg/github.com/gaissmai/go-inet/v2

## License

go-inet is licensed under the MIT License.

