## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DIGITALOCEAN`
along with your [DigitalOcean Personal Access Token Token](https://cloud.digitalocean.com/account/api/tokens).

Example:

{% code title="creds.json" %}
```json
{
  "mydigitalocean": {
    "TYPE": "DIGITALOCEAN",
    "token": "your-digitalocean-token"
  }
}
```
{% endcode %}

The [creds.json](../commands/creds-json.md#example-commands) page in the docs explains how you can generate this dynamically so you can pull the secret token from 1Password or the vault of your choosing.

## Metadata

This provider does not recognize any special metadata fields unique to DigitalOcean.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DIGITALOCEAN = NewDnsProvider("mydigitalocean");

D("example.com", REG_NONE, DnsProvider(DSP_DIGITALOCEAN),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

## Activation

- [Create Personal Access Token](https://cloud.digitalocean.com/account/api/tokens)
- [How to Create a Personal Access Token (documentation)](https://docs.digitalocean.com/reference/api/create-personal-access-token/)

Your access token must have access to create, read, update and delete domain records.

## Supported record types

The [API reference](https://docs.digitalocean.com/reference/api/digitalocean/#tag/Domain-Records) states that these record types are supported:

| Name  | Description |
| ----- | ----------- |
| A     | This record type is used to map an IPv4 address to a hostname. |
| AAAA  | This record type is used to map an IPv6 address to a hostname. |
| CAA   | As specified in RFC-6844, this record type can be used to restrict which certificate authorities are permitted to issue certificates for a domain. |
| CNAME | This record type defines an alias for your canonical hostname (the one defined by an A or AAAA record). |
| MX    | This record type is used to define the mail exchanges used for the domain. |
| NS    | This record type defines the name servers that are used for this zone. |
| TXT   | This record type is used to associate a string of text with a hostname, primarily used for verification. |
| SRV   | This record type specifies the location (hostname and port number) of servers for specific services. |

## Unsupported record types

This means that `ALIAS`, `DHCID`, `DNAME`, `DS`, `FRAME`, `HTTPS`, `LOC`, `OPENPGPKEY`, `PTR`, `SMIMEA`, `SSHFP`, `SVCB`, `TLSA`, `URL`, or `URL301` presumably **do not work** with Digital Ocean.

In 2025, the provider maintainer has confirmed that `ALIAS` and `LOC` records are rejected. The other ones that do not work are expected in this circumstance. `SPF` records are not a problem since they are turned into `TXT` record types.

Since `SOA` record support is so limited we do not provide the option to update it.

## Limitations

- Digitalocean DNS doesn't support `;` value with CAA-records ([DigitalOcean documentation](https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records/))
- While Digitalocean DNS supports TXT records with multiple strings,
  their length is limited by the max API request of 512 octets.
