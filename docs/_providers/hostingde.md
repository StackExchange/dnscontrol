---
name: hosting.de
title: hosting.de Provider
layout: default
jsId: hostingde
---
# hosting.de Provider

## Configuration

In your credentials file, you must provide your [`authToken` and optionally an `ownerAccountId`](https://www.hosting.de/api/#requests-and-authentication).

**If you want to use this provider with http.net or a demo system you need to provide a custom `baseURL`.**

* hosting.de (default): `https://secure.hosting.de`
* http.net: `https://partner.http.net`
* Demo: `https://demo.routing.net`

```json
{
  "hosting.de": {
    "authToken": "YOUR_API_KEY"
  },
  "http.net": {
    "authToken": "YOUR_API_KEY",
    "baseURL": "https://partner.http.net"
  }
}
```

## Usage

Example JavaScript:

```js
var REG_HOSTINGDE = NewRegistrar('hosting.de', 'HOSTINGDE')
var DNS_HOSTINGDE = NewDnsProvider('hosting.de' 'HOSTINGDE');

D('example.tld', REG_HOSTINGDE, DnsProvider(DNS_HOSTINGDE),
    A('test', '1.2.3.4')
);
```

## Customize nameservers

hosting.de has the concept of *nameserver sets* but this provider does not implement it.
The `HOSTINGDE` provider **ignores the default nameserver set** defined in your account!
Instead, it uses hosting.de's nameservers (`ns1.hosting.de.`, `ns2.hosting.de.`, and `ns3.hosting.de.`) by default, regardless of your account settings.

If you want to change this behaviour to, for example, use http.net's nameservers, you can do this by setting an array of strings called `default_ns` in the provider metadata:

```js
var DNS_HTTPNET = NewDnsProvider('http.net', 'HOSTINGDE', {
  default_ns: [
    'ns1.routing.net.',
    'ns2.routing.net.',
    'ns3.routing.net.',
  ],
});
```
