/*
Package inet represents IP-addresses and IP-Blocks as comparable types.

The IP addresses and blocks from this package can be used as keys in maps, freely copied and fast sorted
without prior conversion from/to IPv4/IPv6.

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

Blocks are IP-networks or IP-ranges, e.g.

 192.168.0.1/24              // CIDR, network
 ::1/128                     // CIDR, network
 10.0.0.3-10.0.17.134        // range
 2001:db8::1-2001:db8::f6    // range

*/
package inet
