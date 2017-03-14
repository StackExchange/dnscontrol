# zone2dnscontrol -- Converts a standard DNS zonefile into a DNS zone.

This script helps convert an old-style DNS zone file into the
DNSControl language.  It isn't perfect but it will do 99 percent
of the work so you can focus on just fine-tuning it.

You must give the script both the zone name (i.e. "stackoverflow.com")
and the filename of the zonefile to read.

Output is sent to stdout.

Example:

"""
./convert stackoverflow.com zone.stackoverflow.com
"""

Caveats:

* TTLs are stripped out and/or ignored.
* `$INCLUDE` may not be handled correctly if you are not in the right directory.
* `$GENERATE` is not handled at all.
