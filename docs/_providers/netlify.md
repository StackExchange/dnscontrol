---
name: Netlify
title: Netlify Provider
layout: default
jsId: NETLIFY
---
# Netlify Provider
## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `NETLIFY`
along with a Netlify account personal access token. You can also optionally add an
account slug. This is _typically_ your username on Netlify.

Examples:

```json
{
  "netlify": {
    "TYPE": "NETLIFY",
    "token": "your-netlify-account-access-token",
    "slug": "account-slug" // this is optional
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Netlify.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NETLIFY = NewRegistrar("netlify");
var DSP_NETLIFY = NewDnsProvider("netlify");

D("example.tld", REG_NETLIFY, DnsProvider(DSP_NETLIFY),
    A("test", "1.2.3.4")
);
```

## Activation
DNSControl depends on a Netlify account personal access token.

## Caveats
Empty MX records are not supported.


