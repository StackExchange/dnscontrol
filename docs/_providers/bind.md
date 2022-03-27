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

```json
{
  "bind": {
    "_PROVIDER": "BIND",
    "directory": "myzones",
    "filenameformat": "%U.zone"      << The default
  }
}
```


The BIND accepts some optional metadata via your DNS config when you create the provider:

In this example we set the default SOA settings and NS records.

```js
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
```

# FYI: SOA Records

SOA records are a bit weird in DNSControl.   Most providers auto-generate SOA records and do not permit any modifications. BIND is unique in that it requires users to manage the SOA records themselves.

Because BIND is unique, BIND's SOA support is kind of a hack.  It leaves the SOA record alone, with 2 exceptions:

1. The serial number: If something in the zone changes, the serial number is incremented (see below).
2. Missing SOAs: If there is no SOA record in a zone (or the zone is being created for the first time), the SOA is created.  The initial values are taken from the `default_soa` settings.

The `default_soa` values are only used when creating an SOA for the first time. The values are not used to update an SOA.  *Therefore, the only way to change an existing SOA is to edit the zone file.*

There is an effort to make SOA records handled like A, CNAME, and other records.  See https://github.com/StackExchange/dnscontrol/issues/1131


# FYI: SOA serial numbers

DNSControl tries to maintain the serial number as yyyymmddvv. The algorithm for increasing the serial number is to select the max of (current serial + 1) and (yyyymmdd00). If you use a number larger than today's date (say, 2099000099) DNSControl will simply increment it forever.

The good news is that DNSControl is smart enough to only increment a zone's serial number if something in the zone changed. It does not increment the serial number just because DNSControl ran.

DNSControl does not handle special serial number math such as "looping through zero" nor does it pay attention to the rules around the maximum delta permitted. Those are simply avoided because yyyymmdd99 fits in the first quadrant of the 32-bit serial number space. If you don't understand this paragraph consider yourself lucky; with DNSControl you don't need to.


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

```bash
dnscontrol get-zones --format=nameonly - BIND all
```

If `filenameformat` is defined, `dnscontrol` makes an guess at which
filenames are zones but doesn't try to hard to get it right, which is
mathematically impossible in all cases.  Feel free to file an issue if
your format string doesn't work. I love a challenge!

