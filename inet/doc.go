/*
Package inet represents IP-addresses and IP-Blocks as comparable types.

The IP addresses and blocks from this package can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

This is a package for system programming, all fields are public for easy and fast serialization without special treatment. Anyway, you should not direct modify the fields and bytes, unless you know what you are doing.

IP addresses are represented as fixed arrays of 17 bytes, this ensures natural sorting (IPv4 < IPv6).

 type IP [17]byte

  IP[0]  = version information (4 or 6)
  IP[1:] = IPv4 address or IPv6 address

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

*/
package inet
