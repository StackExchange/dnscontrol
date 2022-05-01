---
name: hosting.de
title: hosting.de Provider
layout: default
jsId: HOSTINGDE
---

# hosting.de Provider

## Configuration

In your credentials file, you must provide your [`authToken` and optionally an `ownerAccountId`](https://www.hosting.de/api/#requests-and-authentication).

### Example `creds.json`

```json
{
  "hosting.de": {
    "TYPE": "HOSTINGDE",
    "authToken": "YOUR_API_KEY"
  }
}
```

## Usage

### Example `dnsconfig.js`

```js
var REG_HOSTINGDE = NewRegistrar('hosting.de', 'HOSTINGDE')
var DNS_HOSTINGDE = NewDnsProvider('hosting.de' 'HOSTINGDE');

D('example.tld', REG_HOSTINGDE, DnsProvider(DNS_HOSTINGDE),
    A('test', '1.2.3.4')
);
```

## Using this provider with http.net and others

http.net and other DNS service providers use an API that is compatible with hosting.de's API.
Using them requires setting the `baseURL` and (optionally) overriding the default nameservers.

### Example http.net configuration

#### Example `creds.json`

```json
{
  "http.net": {
    "TYPE": "HOSTINGDE",
    "authToken": "YOUR_API_KEY",
    "baseURL": "https://partner.http.net"
  }
}
```

#### Example `dnsconfig.js`

```js
var REG_HTTPNET = NewRegistrar('http.net', 'HOSTINGDE');

var DNS_HTTPNET = NewDnsProvider('http.net', 'HOSTINGDE', {
  default_ns: [
    'ns1.routing.net.',
    'ns2.routing.net.',
    'ns3.routing.net.',
  ],
});
```

#### Why this works

hosting.de has the concept of _nameserver sets_ but this provider does not implement it.
The `HOSTINGDE` provider **ignores the default nameserver set** defined in your account to avoid unintentional changes and consolidate the full configuration in DNSControl.
Instead, it uses hosting.de's nameservers (`ns1.hosting.de.`, `ns2.hosting.de.`, and `ns3.hosting.de.`) by default, regardless of your account settings.
Using the `default_ns` metadata, the default nameserver set can be overwritten.
