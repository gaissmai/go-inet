/*
Package inet represents IP-addresses and IP-Blocks as comparable types.

Some missing utility functions in the standard library for IP-addresses and IP-blocks are provided.

This IP representation is comparable and can be sorted very quickly
without prior conversions to/from the different IP versions.

The library is mainly intended for fast ACL-lookups and for IP address management (IPAM) in global scope
and not for host related systems programming.

So, no IP address zone indices are supported and IPv4-mapped IPv6 addresses are stripped down to plain IPv4 addresses.
The information of the prior mapping is discarded.

Blocks are IP-networks or arbitrary IP-ranges, e.g.

 192.168.0.1/24              // network
 ::1/128                     // network
 10.0.0.3-10.0.17.134        // range
 2001:db8::1-2001:db8::f6    // range


*/
package inet
