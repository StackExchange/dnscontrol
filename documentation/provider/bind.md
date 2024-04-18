This provider maintains a directory with a collection of .zone files
as appropriate for ISC BIND, and other systems that use the RFC 1035
zone-file format.

This provider does not generate or update the named.conf file, nor does it deploy the .zone files to the BIND master.
Both of those tasks are different at each site, so they are best done by a locally-written script.


## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `BIND`.

Optional fields include:

* `directory`: Location of the zone files.  Default: `zones` (in the current directory).
* `filenameformat`: The formula used to generate the zone filenames. The default is usually sufficient.  Default: `"%U.zone"`

Example:

{% code title="creds.json" %}
```json
{
  "bind": {
    "TYPE": "BIND",
    "directory": "myzones"
  }
}
```
{% endcode %}

## Meta configuration

This provider accepts some optional metadata in the NewDnsProvider() call.

* `default_soa`: If no SOA record exists in a zone file, one will be created. The values of the new SOA are specified here.
* `default_ns`: Inject these NS records into the zone.

In this example we set the default SOA settings and NS records.

{% code title="dnsconfig.js" %}
```javascript
var DSP_BIND = NewDnsProvider("bind", {
    "default_soa": {
        "master": "ns1.example.com.",
        "mbox": "spamtrap.example.com.",
        "refresh": 3600,
        "retry": 600,
        "expire": 604800,
        "minttl": 1440,
    },
    "default_ns": [
        "ns1.example.com.",
        "ns2.example.com.",
        "ns3.example.com.",
        "ns4.example.com."
    ]
})
```
{% endcode %}

# FYI: SOA Records

SOA records are a bit weird in DNSControl.   Most providers auto-generate SOA records and do not permit any modifications. BIND is unique in that it requires users to manage the SOA records themselves.

Because BIND is unique, BIND's SOA support is kind of a hack.  It leaves the SOA record alone, with 2 exceptions:

1. The serial number: If something in the zone changes, the serial number is incremented (see below).
2. Missing SOAs: If there is no SOA record in a zone (or the zone is being created for the first time), the SOA is created.  The initial values are taken from the `default_soa` settings.

The `default_soa` values are only used when creating an SOA for the first time. The values are not used to update an SOA.  Most people edit the SOA values by manually editing the zonefile or using the `SOA()` function.


# FYI: SOA serial numbers

DNSControl maintains beautiful zone serial numbers.

DNSControl tries to maintain the serial number as yyyymmddvv. The algorithm for increasing the serial number is to select the max of (current serial + 1) and (yyyymmdd00). If you use a number larger than today's date (say, 2099000099) DNSControl will simply increment it forever.

The good news is that DNSControl is smart enough to only increment a zone's serial number if something in the zone changed. It does not increment the serial number just because DNSControl ran.

DNSControl does not handle special serial number math such as "looping through zero" nor does it pay attention to the rules around the maximum delta permitted. Those are simply avoided because yyyymmdd99 fits in the first quadrant of the 32-bit serial number space. If you don't understand this paragraph consider yourself lucky; with DNSControl you don't need to.


# filenameformat

The `filenameformat` parameter specifies the file name to be used when
writing the zone file. The default (`%U.zone`) is acceptable in most cases: the
file name is the name as specified in the `D()` function plus ".zone".

The filenameformat is a string with a few printf-like `%` verbs:

  * `%U`  the domain name as specified in `D()`
  * `%D`  the domain name without any split horizon tag (the "example.com" part of "example.com!tag")
  * `%T`  the split horizon tag, or "" (the "tag" part of "example.com!tag")
  * `%?x` this returns `x` if the split horizon tag is non-null, otherwise nothing. `x` can be any printable.
  * `%%`  `%`
  * ordinary characters (not `%`) are copied unchanged to the output stream
  * FYI: format strings must not end with an incomplete `%` or `%?`

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
`D("example.com!inside")` and `D("example.com!outside")`.  This
assumes two BIND providers are configured in `creds.json`, each with
a different `directory` setting. Otherwise `dnscontrol` will write
both domains to the same file, flapping between the two back and
forth.

(new in v4.2.0) `dnscontrol push` will create subdirectories along the path to
the filename. This includes both the portion of the path created by the
`directory` setting and the `filenameformat` setting. The automatic creation of
subdirectories is disabled if `dnscontrol` is running as root for security
reasons.

# FYI: get-zones

The DNSControl `get-zones all` subcommand scans the directory for
any files named `*.zone` and assumes they are zone files.

```shell
dnscontrol get-zones --format=nameonly - BIND all
```

If `filenameformat` is defined, `dnscontrol` makes an guess at which
filenames are zones but doesn't try to hard to get it right, which is
mathematically impossible in some cases.  Feel free to file an issue if
your format string doesn't work. I love a challenge!
