---
name: Gcore
title: Gcore Provider
layout: default
jsId: GCORE
---
# Gcore Provider
## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `GCORE`
along with a Gcore account API token.

Example:

```json
{
  "gcore": {
    "TYPE": "GCORE",
    "api-key": "your-gcore-api-key"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Gcore.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("none");    // No registrar.
var DSP_GCORE = NewDnsProvider("gcore");  // Gcore

D("example.tld", REG_NONE, DnsProvider(DSP_GCORE),
    A("test", "1.2.3.4")
);
```

## Activation

DNSControl depends on a Gcore account API token.

You can obtain your API token on this page: <https://accounts.gcore.com/profile/api-tokens>
