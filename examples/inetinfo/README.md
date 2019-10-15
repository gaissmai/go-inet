# inetinfo

Print useful information about IP addresses, CIDRs and ranges.

## Usage

```bash
inetinfo ip_or_CIDR_or_range
```

Example:

```bash
inetinfo 2001:db8:c::/116

Version:   6
Prefix:    2001:db8:c::/116
Base:      2001:db8:c::
Last:      2001:db8:c::fff
Mask:      ffff:ffff:ffff:ffff:ffff:ffff:ffff:f000
Wildcard:  ::fff
Bits:      12 bits
Size:      4096 addrs
```

Example:

```bash
inetinfo 2001:db8:c::

Version:   6
RFC:       2001:db8:c::
Expand:    2001:0db8:000c:0000:0000:0000:0000:0000
Reverse:   0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.c.0.0.0.8.b.d.0.1.0.0.2
```
