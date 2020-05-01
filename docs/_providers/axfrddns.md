---
name: AXFRDDNS
title: AXFR+DDNS Provider
layout: default
jsId: AXFRDDNS
---
# AXFR+DDNS Provider

This provider is able to work with any authoritative DNS server accepting AXFR requests (RFC5936) and Dynamic Update (RFC2136).

It has been tested with [BIND](https://www.isc.org/bind/), [Knot](https://www.knot-dns.cz/), and [Yadifa](https://www.yadifa.eu/home).

## Configuration

### Authentication

The AXFR+DDNS provider might work without anything in `creds.json` if the primary master of the zone accepts transfers and updates without TSIG authentication. But for non widely-open server, the authentication keys should be provided in `creds.json`.

{% highlight json %}
{
    "axfrddns": {
        "transfer-key": "hmac-sha256:transfer-key-id:Base64EncodedSecret=",
        "update-key": "hmac-sha256:update-key-id:AnotherSecret="
    }
}
{% endhighlight %}

The `transfer-key` will be used to authenticate AXFR request, and the `update-key` will be used to authenticate the Dynamic Updates. Both keys are optional, and you could provide only one them.

If distinct zones requires distinct keys, you might instantiate the provider multiple times:

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

The AXFR+DDNS provider can be configured with a list of default nameservers. They will be added to all the zones handled by the provider.

This list can be provided either as metadata or in `creds.json`. Only the later allows `get-zones` to work properly.

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

By default, the AXFR+DDNS provider will send the AXFR requests and the updates to the first nameserver of the zone, usually known as the "primary master". Typically, this is the first of the default nameservers. Though, on some networks, the primary master is a private node, hidden behind slaves, and it does not appear in the `NS` records of the zone. In that case, the IP or the name of the primary server must be provided in `creds.json`. With this option, a non-standard port might be used.

{% highlight json %}
{
   master = "10.20.30.40:5353"
}
{% endhighlight %}

When no nameserver appears in the zone, and no default nameservers nor custom master are configured, the AXFR+DDNS provider will fail.

## FYI: get-zones

When using `get-zones`, a custom master or a list of default nameservers should be configured in `creds.json`.

THe AXFR+DDNS provider does not display DNSSec Records. But, if any DNSSec records is found in the zone, it will replace all of them with a single placeholder record:

{% highlight %}
__dnssec         IN TXT   "Domain has DNSSec records, not displayed here."
{% endhighlight %}

## FYI: create-domain

The AXFR+DDNS provider is not able to create domain.

## FYI: AUTODNSSEC

The AXFR+DDNS provider is not able to ask the DNS server to sign the zone. But, it is able to check whether the server seems to do so or not.

When AutoDNSSEC is set, the AXFR+DDNS provider will emit a warning when no RRSIG, DNSKEY or NSEC records are found in the zone.

When AutoDNSSEC is not set, the AXFR+DDNS provider will emit a warning when RRSIG, DNSKEY or NSEC records are found in the zone.
