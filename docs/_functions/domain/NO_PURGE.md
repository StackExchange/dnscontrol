---
name: NO_PURGE
---

NO_PURGE indicates that records should not be deleted from a domain.
Records will be added and updated, but not removed.

NO_PURGE is generally used in very specific situations:

* A domain is managed by some other system and DNSControl is only used to insert a few specific records and/or keep them updated. For example a DNS Zone that is managed by Active Directory, but DNSControl is used to update a few, specific, DNS records. In this case we want to specify the DNS records we are concerned with but not delete all the other records.  This is a risky use of NO_PURGE since, if NO_PURGE was removed (or buggy) there is a chance you could delete all the other records in the zone, which could be a disaster. That said, domains with some records updated using Dynamic DNS have no other choice.
* To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. NO_PURGE will leave those records alone.

In this example DNSControl will insert "foo.example.com" into the
zone, but otherwise leave the zone alone.  Changes to "foo"'s IP
address will update the record. Removing the A("foo", ...) record
from dnscontrol will leave the record in place.

{% include startExample.html %}
{% highlight js %}
D("example.com", .... , NO_PURGE,
  A("foo","1.2.3.4")
);
{%endhighlight%}
{% include endExample.html %}

The main caveat of NO_PURGE is that intentionally deleting records
becomes more difficult. Suppose a NO_PURGE zone has an record such
as A("ken", "1.2.3.4"). Removing the record from dnsconfig.js will
not delete "ken" from the domain. DNSControl has no way of knowing
the record was deleted from the file  The DNS record must be removed
manually.  Users of NO_PURGE are prone to finding themselves with
an accumulation of orphaned DNS records. That's easy to fix for a
small zone but can be a big mess for large zones.

Not all providers support NO_PURGE. For example the BIND provider
rewrites zone files from scratch each time, which precludes supporting
NO_PURGE.  DNSControl will exit with an error if NO_PURGE is used
on a driver that does not support it.

There is also `PURGE` command for completeness. `PURGE` is the
default, thus this command is a no-op.
