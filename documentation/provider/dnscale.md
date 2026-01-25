## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DNSCALE`
along with your DNScale API key.

Example:

{% code title="creds.json" %}
```json
{
  "dnscale": {
    "TYPE": "DNSCALE",
    "api_key": "dnscale_your-api-key-here"
  }
}
```
{% endcode %}

## Metadata

This provider does not recognize any special metadata fields unique to DNScale.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DNSCALE = NewDnsProvider("dnscale");

D("example.com", REG_NONE, DnsProvider(DSP_DNSCALE),
    A("@", "192.0.2.1"),
    A("www", "192.0.2.1"),
    AAAA("@", "2001:db8::1"),
    CNAME("blog", "example.github.io."),
    MX("@", 10, "mail.example.com."),
    TXT("@", "v=spf1 include:_spf.google.com ~all"),
    CAA("@", "issue", "letsencrypt.org"),
END);
```
{% endcode %}

## Activation

DNScale requires an API key which can be obtained from your [DNScale dashboard](https://app.dnscale.eu/dashboard).

## Supported Record Types

DNScale supports the following record types:

- A
- AAAA
- ALIAS
- CAA
- CNAME
- HTTPS
- MX
- NS
- PTR
- SRV
- SSHFP
- SVCB
- TLSA
- TXT

## New domains

If a domain does not exist in your DNScale account, DNSControl will automatically create it when you run `dnscontrol push`.

## API Documentation

For more information about the DNScale API, see the [DNScale API documentation](https://dnscale.eu/api/overview).
