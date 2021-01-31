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
    "filenameformat": "%T%U%D.zone"      << The default
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
writing the zone file. The default is acceptable in most cases.
The `dnscontrol get-zones` command only scans for filenames in the
default format.

The filenameformat is a string with a few printf-like `%` directives:

  * `%U`  the domain name as specified in dnsconfig.js
  * `%D`  the domain name, stripped of any tags
  * `%T`  the split horizon tag, see `D()` for info
  * `%*x`  returns `x` if tag is non-null, otherwise nothing. `x` can be any printable.
  * `%%`  `%`
  * ordinary characters (not `%`), are copied unchanged to the output stream.
  * `%` may not be the last char in a string

Typical values:

  * "%T%*U%D.zone"  (the default) Ex: `tag_example.com.zone` or `example.com.zone`
  * "db_%D"  Ex: `db_example.com` (assumes no tags)
  * "db_%T%U%D"  Ex: `db_inside_example.com` or `db_example.com`

# FYI: get-zones

The dnscontrol `get-zones all` subcommand scans the directory for
any files named `*.zone` and assumes they are zone files.

If `filenameformat` is defined, the code makes a simple guess
at filenames. It isn't a reliable algorithm, but feel free to
file an issue if your format string doesn't work.

```
dnscontrol get-zones --format=nameonly - BIND all
```

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
