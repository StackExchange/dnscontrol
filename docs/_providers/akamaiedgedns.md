---
name: AkamaiEdgeDns
title: Akamai Edge DNS Provider
layout: default
jsId: AKAMAIEDGEDNS
---

# Akamai Edge DNS Provider

"Akamai Edge DNS Provider" configures Akamai's
[Edge DNS](https://www.akamai.com/us/en/products/security/edge-dns.jsp) service.

This provider interacts with Edge DNS via the
[Edge DNS Zone Management API](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html).

Before you can use this provider, you need to create an "API Client" with authorization to use the
[Edge DNS Zone Management API](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html).

See the "Get Started" section of [Edge DNS Zone Management API](https://developer.akamai.com/api/cloud_security/edge_dns_zone_management/v2.html),
which says, "To enable this API, choose the API service named DNS—Zone Record Management, and set the access level to READ-WRITE."

Follow directions at [Authenticate With EdgeGrid](https://developer.akamai.com/getting-started/edgegrid) to generate
the required credentials.

## Configuration

In the credentials file (creds.json), you must provide the following:

```json
"akamaiedgedns": {
    "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "host": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.xxxx.akamaiapis.net",
    "access_token": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "client_token": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "contract_id": "X-XXXX",
    "group_id": "NNNNNN"
}
```

## Usage

A new zone created by DNSControl:

```bash
dnscontrol create-domains
```

automatically creates SOA and authoritative NS records.

Akamai assigns a unique set of authoritative nameservers for each contract.  These authorities should be
used as the NS records on all zones belonging to this contract.

The NS records for these authorities have a TTL of 86400.

Add:

```js
NAMESERVER_TTL(86400)
```

modifier to the dnscontrol.js D() function so that DNSControl does not change the TTL of the authoritative NS records.

Example 'dnsconfig.js':

```js
var REG_NONE = NewRegistrar('none', 'NONE');
var DNS_AKAMAIEDGEDNS = NewDnsProvider('akamaiedgedns', 'AKAMAIEDGEDNS');

D('example.com', REG_NONE, DnsProvider(DNS_AKAMAIEDGEDNS),
  NAMESERVER_TTL(86400),
  AUTODNSSEC_ON,
  AKAMAICDN("@", "www.preconfigured.edgesuite.net", TTL(20)),
  A('foo','1.2.3.4')
);
```

AKAMAICDN is a proprietary record type that is used to configure [Zone Apex Mapping](https://blogs.akamai.com/2019/08/fast-dns-zone-apex-mapping-dnssec.html).
The AKAMAICDN target must be preconfigured in the Akamai network.
