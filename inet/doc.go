/*
Package inet represents IP-addresses and IP-Blocks as comparable types.

They can be used as keys in maps, freely copied and fast sorted
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

Blocks are represented as a struct of three IP addresses:

 type Block struct {
  Base IP
  Last IP
  Mask IP  // may be zero for begin-end ranges
 }

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

This is a package for system programming, all fields are public for easy and fast serialization without special treatment.
Anyway, you should not direct modify the fields and bytes, unless you know what you are doing.
*/
package inet
