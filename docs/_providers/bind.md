---
name: BIND
title: BIND Provider
layout: default
jsId: BIND
---
# BIND Provider
This provider maintains a directory with a collection of .zone files.

This provider does not generate or update the named.conf file, nor does it deploy the .zone files to the BIND master.
Both of those tasks are different at each site, so they are best done by a locally-written script.


## Configuration
The BIND provider does not require anything in `creds.json`. However
you can specify a `directory` where the provider will look for and create zone files. The default is the `zones` directory (in the current directory).

{% highlight json %}
{
  "bind": {
    "directory": "myzones",
    "filenameformat": "%U.zone"      << The default
  }
}
{% endhighlight %}


The BIND accepts some optional metadata via your DNS config when you create the provider:

In this example we set the default SOA settings and NS records.

{% highlight javascript %}
var BIND = NewDnsProvider('bind', 'BIND', {
        'default_soa': {
        'master': 'ns1.example.tld.',
        'mbox': 'sysadmin.example.tld.',
        'refresh': 3600,
        'retry': 600,
        'expire': 604800,
        'minttl': 1440,
    },
    'default_ns': [
        'ns1.example.tld.',
        'ns2.example.tld.',
        'ns3.example.tld.',
        'ns4.example.tld.'
    ]
})
{% endhighlight %}

# filenameformat

The `filenameformat` parameter specifies the file name to be used when
writing the zone file. The default is acceptable in most cases: the
name as specified in the `D()` function, plus ".zone".

The filenameformat is a string with a few printf-like `%` verbs:

  * `%U`  the domain name as specified in `D()`
  * `%D`  the domain name without any split horizon tag
  * `%T`  the split horizon tag, or "", see `D()`
  * `%?x` this returns `x` if the split horizon tag is non-null, otherwise nothing. `x` can be any printable.
  * `%%`  `%`
  * ordinary characters (not `%`) are copied unchanged to the output stream
  * FYI: format strings must not end with an incomplete `%` or `%?`
  * FYI: `/` or other filesystem separators result in undefined behavior

Typical values:

  * `%U.zone` (The default)
    * `example.com.zone` or `example.com!tag.zone`
  * `%T%*U%D.zone`  (optional tag and `_` + domain + `.zone`)
    * `tag_example.com.zone` or `example.com.zone`
  * `db_%T%?_%D`
    * `db_inside_example.com` or `db_example.com`
  * `db_%D`
    * `db_example.com`

The last example will generate the same name for both
`D("example.tld!inside")` and `D("example.tld!outside")`.  This
assumes two BIND providers are configured in `creds.json`, eacch with
a different `directory` setting. Otherwise `dnscontrol` will write
both domains to the same file, flapping between the two back and
forth.

# FYI: get-zones

The dnscontrol `get-zones all` subcommand scans the directory for
any files named `*.zone` and assumes they are zone files.

```
dnscontrol get-zones --format=nameonly - BIND all
```

If `filenameformat` is defined, `dnscontrol` makes an guess at which
filenames are zones but doesn't try to hard to get it right, which is
mathematically impossible in all cases.  Feel free to file an issue if
your format string doesn't work. I love a challenge!

# FYI: SOA Records

DNSControl assumes that SOA records are managed by the provider.  Most
providers simply generate the SOA record for you and do not permit you
to control it at all.  The BIND provider is unique in that it must emulate
what most DNS-as-a-service providers do.

When DNSControl reads a BIND zonefile:

* If there was no SOA record, one is created using the `default_soa`
  settings listed above.
* When generating a new zonefile, the SOA serial number is
  updated.

DNSControl tries to maintain the serial number as yyyymmddvv. If the
existing serial number is significantly higher it will simply
increment the value by 1.

If you need to edit the SOA fields, the best way is to edit the
zonefile directly, then run `dnscontrol preview` and `dnscontrol push`
as normal.
