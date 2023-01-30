This provider uses the native DNS protocols. It uses the AXFR (RFC5936,
Zone Transfer Protocol) to retrieve the existing records and DDNS
(RFC2136, Dynamic Update) to make corrections. It can use TSIG (RFC2845) or
IP-based authentication (ACLs).

It is able to work with any standards-compliant
authoritative DNS server. It has been tested with
[BIND](https://www.isc.org/bind/), [Knot](https://www.knot-dns.cz/),
and [Yadifa](https://www.yadifa.eu/home).

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AXFRDDNS`.

### Connection modes

Zone transfers default to TCP, DDNS updates default to UDP when
using this provider.

The following two parameters in `creds.json` allow switching
to TCP or TCP over TLS.

* `update-mode`: May contain `udp` (the default), `tcp`, or `tcp-tls`.
* `transfer-mode`: May contain `tcp` (the default), or `tcp-tls`.

### Authentication

Authentication information is included in the `creds.json` entry for
the provider:

* `transfer-key`: If this exists, the value is used to authenticate AXFR transfers.
* `update-key`: If this exists, the value is used to authenticate DDNS updates.

For instance, your `creds.json` might looks like:

```json
{
  "axfrddns": {
    "TYPE": "AXFRDDNS",
    "transfer-key": "hmac-sha256:transfer-key-id:Base64EncodedSecret=",
    "update-key": "hmac-sha256:update-key-id:AnotherSecret="
  }
}
```

If either key is missing, DNSControl defaults to IP-based ACL
authentication for that function. Including both keys is the most
secure option. Omitting both keys defaults to IP-based ACLs for all
operations, which is the least secure option.

If distinct zones require distinct keys, you will need to instantiate the
provider once for each key:

```javascript
var DSP_AXFRDDNS_A = NewDnsProvider("axfrddns-a");
var DSP_AXFRDDNS_B = NewDnsProvider("axfrddns-b");
```

And update `creds.json` accordingly:

```json
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
```

### Default nameservers

The AXFR+DDNS provider can be configured with a list of default
nameservers. They will be added to all the zones handled by the
provider.

This list can be provided either as metadata or in `creds.json`. Only
the later allows `get-zones` to work properly.

```javascript
var DSP_AXFRDDNS = NewDnsProvider("axfrddns", {
        "default_ns": [
            "ns1.example.tld.",
            "ns2.example.tld.",
            "ns3.example.tld.",
            "ns4.example.tld."
        ]
    }
}
```

```json
{
   nameservers = "ns1.example.tld,ns2.example.tld,ns3.example.tld,ns4.example.tld"
}
```

### Primary master

By default, the AXFR+DDNS provider will send the AXFR requests and the
DDNS updates to the first nameserver of the zone, usually known as the
"primary master". Typically, this is the first of the default
nameservers. Though, on some networks, the primary master is a private
node, hidden behind slaves, and it does not appear in the `NS` records
of the zone. In that case, the IP or the name of the primary server
must be provided in `creds.json`. With this option, a non-standard
port might be used.

```json
{
   master = "10.20.30.40:5353"
}
```

When no nameserver appears in the zone, and no default nameservers nor
custom master are configured, the AXFR+DDNS provider will fail with
the following error message:

```text
[Error] AXFRDDNS: the nameservers list cannot be empty.
Please consider adding default `nameservers` or an explicit `master` in `creds.json`.
```


## Server configuration examples

### Bind9

Here is a sample `named.conf` example for an authauritative server on
zone `example.tld`. It uses a simple IP-based ACL for the AXFR
transfer and a conjunction of TSIG and IP-based ACL for the updates.

```javascript
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
```

## FYI: get-zones

When using `get-zones`, a custom master or a list of default
nameservers should be configured in `creds.json`.

THe AXFR+DDNS provider does not display DNSSec records. But, if any
DNSSec records is found in the zone, it will replace all of them with
a single placeholder record:

```text
    __dnssec         IN TXT   "Domain has DNSSec records, not displayed here."
```

## FYI: create-domain

The AXFR+DDNS provider is not able to create domain.

## FYI: AUTODNSSEC

The AXFR+DDNS provider is not able to ask the DNS server to sign the zone. But, it is able to check whether the server seems to do so or not.

When AutoDNSSEC is enabled, the AXFR+DDNS provider will emit a warning when no RRSIG, DNSKEY or NSEC records are found in the zone.

When AutoDNSSEC is disabled, the AXFR+DDNS provider will emit a warning when RRSIG, DNSKEY or NSEC records are found in the zone.

When AutoDNSSEC is not enabled or disabled, no checking is done.

## FYI: MD5 Support

By default the used DNS Go package by miekg has deprecated supporting the (insecure) MD5 algorithm [https://github.com/miekg/dns/commit/93945c284489394b77653323d11d5de83a2a6fb5](https://github.com/miekg/dns/commit/93945c284489394b77653323d11d5de83a2a6fb5). Some providers like the Leibniz Supercomputing Centre (LRZ) located in Munich still use this algorithm to authenticate internal dynamic DNS updates. To compensate the lack of MD5 a custom MD5 TSIG Provider was added into DNSControl.
