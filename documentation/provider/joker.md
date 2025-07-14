# Joker DNS Provider

## Configuration

To use this provider, add an entry to `creds.json` with your Joker.com credentials:

{% code title="creds.json" %}
```json
{
  "joker": {
    "TYPE": "JOKER",
    "username": "your-username@joker.com",
    "password": "your-password"
  }
}
```
{% endcode %}

You must have a reseller account in joker.com to use the DMAPI.

Alternatively, you can use an API key (if you have created one on the Joker.com website):

{% code title="creds.json" %}
```json
{
  "joker": {
    "TYPE": "JOKER",
    "api-key": "your-api-key"
  }
}
```
{% endcode %}

## Metadata

This provider does not recognize any special metadata fields unique to Joker.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_JOKER = NewDnsProvider("joker");

D("example.tld", REG_NONE, DnsProvider(DSP_JOKER),
    A("test", "1.2.3.4"),
    CNAME("www", "test"),
    MX("@", 10, "mail.example.tld."),
    TXT("_dmarc", "v=DMARC1; p=reject; rua=mailto:dmarc@example.tld"),
END);
```
{% endcode %}

## Limitations

- This provider updates entire zones, not individual records
- Concurrent operations are not supported due to session-based authentication
- Some record types are not supported (see provider capabilities)
- Minimum TTL is 300 seconds

## Notes

- The provider uses Joker's DMAPI (Domain Management API)
- Authentication uses session-based tokens that expire after inactivity
- Zone updates replace the entire zone content
- The provider supports both username/password and API key authentication

## Supported Record Types

- A
- AAAA
- CNAME
- MX
- TXT
- SRV
- CAA
- NAPTR

## Unsupported Record Types

- ALIAS
- DS
- DNSKEY
- HTTPS
- LOC
- PTR
- SOA
- SSHFP
- SVCB
- TLSA