---
name: AXFRDDNS
title: AXFR+DDNS Provider
layout: default
jsId: AXFRDDNS
---
# AXFR+DDNS Provider

This provider uses the native DNS protocols. It uses the AXFR (RFC5936,
Zone Transfer Protocol) to retrieve the existing records and DDNS
(RFC2136, Dynamic Update) to make corrections. It can use TSIG (RFC2845) or
IP-based authentication (ACLs).

It is able to work with any standards-compliant
authoritative DNS server. It has been tested with
[BIND](https://www.isc.org/bind/), [Knot](https://www.knot-dns.cz/),
and [Yadifa](https://www.yadifa.eu/home).

## Configuration

### Authentication

Authentication information is included in the `creds.json` entry for
the provider:

* `transfer-key`: If this exists, the value is used to authenticate AXFR transfers.
* `update-key`: If this exists, the value is used to authenticate DDNS updates.

For instance, your `creds.json` might looks like:

{% highlight json %}
{
    "axfrddns": {
        "transfer-key": "hmac-sha256:transfer-key-id:Base64EncodedSecret=",
        "update-key": "hmac-sha256:update-key-id:AnotherSecret="
    }
}
{% endhighlight %}

If either key is missing, DNSControl defaults to IP-based ACL
authentication for that function. Including both keys is the most
secure option. Omitting both keys defaults to IP-based ACLs for all
operations, which is the least secure option.

If distinct zones require distinct keys, you will need to instantiate the
provider once for each key:

{% highlight javascript %}
var AXFRDDNS_A = NewDnsProvider('axfrddns-a', 'AXFRDDNS'}
var AXFRDDNS_B = NewDnsProvider('axfrddns-b', 'AXFRDDNS'}
{% endhighlight %}

And update `creds.json` accordingly:

{% highlight json %}
{
    "axfrddns-a": {
        "transfer-key": "hmac-sha256:transfer-key-id:Base64EncodedSecret=",
        "update-key": "hmac-sha256:update-key-id:AnotherSecret="
    },
    "axfrddns-b": {
        "transfer-key": "hmac-sha512:transfer-key-id-B:SmallSecret=",
        "update-key": "hmac-sha512:update-key-id-B:YetAnotherSecret="
    }
}
{% endhighlight %}

### Default nameservers

The AXFR+DDNS provider can be configured with a list of default
nameservers. They will be added to all the zones handled by the
provider.

This list can be provided either as metadata or in `creds.json`. Only
the later allows `get-zones` to work properly.

{% highlight javascript %}
var AXFRDDNS = NewDnsProvider('axfrddns', 'AXFRDDNS',
    'default_ns': [
        'ns1.example.tld.',
        'ns2.example.tld.',
        'ns3.example.tld.',
        'ns4.example.tld.'
    ]
}
{% endhighlight %}

{% highlight json %}
{
   nameservers = "ns1.example.tld,ns2.example.tld,ns3.example.tld,ns4.example.tld"
}
{% endhighlight %}

### Primary master

By default, the AXFR+DDNS provider will send the AXFR requests and the
DDNS updates to the first nameserver of the zone, usually known as the
"primary master". Typically, this is the first of the default
nameservers. Though, on some networks, the primary master is a private
node, hidden behind slaves, and it does not appear in the `NS` records
of the zone. In that case, the IP or the name of the primary server
must be provided in `creds.json`. With this option, a non-standard
port might be used.

{% highlight json %}
{
   master = "10.20.30.40:5353"
}
{% endhighlight %}

When no nameserver appears in the zone, and no default nameservers nor
custom master are configured, the AXFR+DDNS provider will fail with
the following error message:

{% highlight text %}
[Error] AXFRDDNS: the nameservers list cannot be empty.
Please consider adding default `nameservers` or an explicit `master` in `creds.json`.
{% endhighlight %}


## Server configuration examples

### Bind9

Here is a sample `named.conf` example for an authauritative server on
zone `example.tld`. It uses a simple IP-based ACL for the AXFR
transfer and a conjunction of TSIG and IP-based ACL for the updates.

{% highlight javascript %}
options {

	listen-on { any; };
	listen-on-v6 { any; };

	allow-query { any; };
	allow-notify { none; };
	allow-recursion { none; };
	allow-transfer { none; };
	allow-update { none; };
	allow-query-cache { none; };

};

zone "example.tld" {
  type master;
  file "/etc/bind/db.example.tld";
  allow-transfer { example-transfer; };
  allow-update { example-update; };
};

## Allow transfer to anyone on our private network

acl example-transfer {
    172.17.0.0/16;
};

## Allow update only from authenticated client on our private network

acl example-update {
  ! {
   !172.17.0.0/16;
   any;
  };
  key update-key-id;
};

key update-key-id {
  algorithm HMAC-SHA256;
  secret "AnotherSecret=";
};
{% endhighlight %}

## FYI: get-zones

When using `get-zones`, a custom master or a list of default
nameservers should be configured in `creds.json`.

THe AXFR+DDNS provider does not display DNSSec records. But, if any
DNSSec records is found in the zone, it will replace all of them with
a single placeholder record:

{% highlight text %}
    __dnssec         IN TXT   "Domain has DNSSec records, not displayed here."
{% endhighlight %}

## FYI: create-domain

The AXFR+DDNS provider is not able to create domain.

## FYI: AUTODNSSEC

The AXFR+DDNS provider is not able to ask the DNS server to sign the zone. But, it is able to check whether the server seems to do so or not.

When AutoDNSSEC is enabled, the AXFR+DDNS provider will emit a warning when no RRSIG, DNSKEY or NSEC records are found in the zone.

When AutoDNSSEC is disabled, the AXFR+DDNS provider will emit a warning when RRSIG, DNSKEY or NSEC records are found in the zone.

When AutoDNSSEC is not enabled or disabled, no checking is done.
