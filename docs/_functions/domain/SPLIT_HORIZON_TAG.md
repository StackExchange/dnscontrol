---
name: SPLIT_HORIZON_TAG
parameters:
  - name
---

`SPLIT_HORIZON_TAG` enables "split horizon DNS". It allows duplicate
domain names to be specified as long as each has a unique tag.

A domain without a `SPLIT_HORIZON_TAG` has the tag "" (null string).
Therefore all but one domains in a split horizon configuration must
have the `SPLIT_HORIZON_TAG()` but it is good form to specify
`SPLIT_HORIZON_TAG` on all domains in a split horizon.

When used, the `--domains=example.com`` command-line flag will match
all domains named `example.com`. If you wish to specify just one
domain, append `!` and the tag name. For example,
`--domains=example.com!external,foo.com` would specify the second
example below, and the domain `foo.com` (not mentioned in the
example).  `--domains=example.com!` will match the domain without a
tag.

Caveat: The output of `preview`, `push` and other `dnscontrol` sub
commands print the domain name without the tag, which may be
confusing.

Example:

{% include startExample.html %}
{% highlight js %}
// Split horizon example.  Here we permit
// duplicate domains, each sending data to different providers.
var DNS_BIND_INTERNAL = NewDnsProvider("bind_internal","BIND");
var DNS_BIND_EXTERNAL = NewDnsProvider("bind_external","BIND");
D("example.com", REGISTRAR,
  DnsProvider(DNS_BIND_INTERNAL),
  A("@","10.2.3.4"),
);
D("example.com", REGISTRAR,
  SPLIT_HORIZON_TAG("external"),  // Differentiate from the other
  DnsProvider(DNS_BIND_EXTERNAL),
  A("@","99.99.99.99"),
);

{%endhighlight%}
{% include endExample.html %}
