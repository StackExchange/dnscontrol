"Akamai Edge DNS Provider" configures Akamai's
[Edge DNS](https://www.akamai.com/products/edge-dns) service.

This provider interacts with Edge DNS via the
[Edge DNS Zone Management API](https://techdocs.akamai.com/edge-dns/reference/edge-dns-api).

Before you can use this provider, you need to create an "API Client" with authorization to use the
[Edge DNS Zone Management API](https://techdocs.akamai.com/edge-dns/reference/edge-dns-api).

See the "Get Started" section of [Edge DNS Zone Management API](https://techdocs.akamai.com/edge-dns/reference/edge-dns-api),
which says, "To enable this API, choose the API service named DNS—Zone Record Management, and set the access level to READ-WRITE."

Follow directions at [Authenticate With EdgeGrid](https://www.akamai.com/developer) to generate
the required credentials.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AKAMAIEDGEDNS` along with the authentication fields.

Example:

{% code title="creds.json" %}
```json
{
  "akamaiedgedns": {
    "TYPE": "AKAMAIEDGEDNS",
    "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "host": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.xxxx.akamaiapis.net",
    "access_token": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "client_token": "akaa-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "contract_id": "X-XXXX",
    "group_id": "NNNNNN"
  }
}
```
{% endcode %}

## Limitations

### Records

#### AKAMAICDN

The AKAMAICDN target must be an Edge Hostname preconfigured in your Akamai account. 

The AKAMAICDN record must have a TTL of 20 seconds.

The AKAMAICDN record may only be used at the zone apex (`@`) if an AKAMAITLC record hasn't been used.

#### AKAMAITLC

The AKAMAITLC record can only be used at the zone apex (`@`).

The AKAMAITLC record can only be used once per zone.

#### ALIAS
Akamai Edge DNS does directly support `ALIAS` records. This provider will convert `ALIAS` records used at the 
zone apex (`@`) to `AKAMAITLC` records, and any other names to `CNAME` records.

### Secondary zones

This provider only supports creating primary zones in Akamai. If a secondary zone has been manually created, only `AKAMAICDN` and `AKAMAITLC` records can be managed, as all other records are read-only.

## Usage

A new primary zone created by DNSControl:

```shell
dnscontrol create-domains
```

automatically creates SOA and authoritative NS records.

Akamai assigns a unique set of authoritative nameservers for each contract.  These authorities should be
used as the NS records on all zones belonging to this contract.

The NS records for these authorities have a TTL of 86400.

Add:

{% code title="dnsconfig.js" %}
```javascript
NAMESERVER_TTL(86400)
```
{% endcode %}

modifier to the dnscontrol.js D() function so that DNSControl does not change the TTL of the authoritative NS records.

Example `dnsconfig.js`:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_AKAMAIEDGEDNS = NewDnsProvider("akamaiedgedns");

D("example.com", REG_NONE, DnsProvider(DSP_AKAMAIEDGEDNS),
  NAMESERVER_TTL(86400),
  AUTODNSSEC_ON,
  AKAMAICDN("@", "preconfigured.edgesuite.net", TTL(20)),
  AKAMAICDN("www", "www.preconfigured.edgesuite.net", TTL(20)),
  A("foo", "1.2.3.4"),
);
```
{% endcode %}
