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

## Feature Flags

<!-- provider-features-start -->
- Provider Type
  - Official Support: ❌
  - DNS Provider: ✅
  - Registrar: ❌
- Provider API
  - [Concurrency Verified](../advanced-features/concurrency-verified.md): ❌
  - [dual host](../advanced-features/dual-host.md): ❔
  - create-domains: ✅
  - get-zones: ✅
- DNS extensions
  - [`ALIAS`](../language-reference/domain-modifiers/ALIAS.md): ✅
  - [`DNAME`](../language-reference/domain-modifiers/DNAME.md): ❔
  - [`LOC`](../language-reference/domain-modifiers/LOC.md): ❌
  - [`PTR`](../language-reference/domain-modifiers/PTR.md): ✅
  - [`SOA`](../language-reference/domain-modifiers/SOA.md): ❔
- Service discovery
  - [`DHCID`](../language-reference/domain-modifiers/DHCID.md): ❔
  - [`NAPTR`](../language-reference/domain-modifiers/NAPTR.md): ❔
  - [`SRV`](../language-reference/domain-modifiers/SRV.md): ✅
  - [`SVCB`](../language-reference/domain-modifiers/SVCB.md): ✅
- Security
  - [`CAA`](../language-reference/domain-modifiers/CAA.md): ✅
  - [`HTTPS`](../language-reference/domain-modifiers/HTTPS.md): ✅
  - [`SMIMEA`](../language-reference/domain-modifiers/SMIMEA.md): ❔
  - [`SSHFP`](../language-reference/domain-modifiers/SSHFP.md): ✅
  - [`TLSA`](../language-reference/domain-modifiers/TLSA.md): ✅
- DNSSEC
  - [`AUTODNSSEC`](../language-reference/domain-modifiers/AUTODNSSEC_ON.md): ❔
  - [`DNSKEY`](../language-reference/domain-modifiers/DNSKEY.md): ❔
  - [`DS`](../language-reference/domain-modifiers/DS.md): ❔
<!-- provider-features-end -->
