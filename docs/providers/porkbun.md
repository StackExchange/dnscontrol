---
name: Porkbun
title: Porkbun Provider
layout: default
jsId: PORKBUN
---
# Porkbun Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `PORKBUN`
along with your `api_key` and `secret_key`. More info about authentication can be found in [Getting started with the Porkbun API](https://kb.porkbun.com/article/190-getting-started-with-the-porkbun-api).

Example:

```json
{
  "porkbun": {
    "TYPE": "PORKBUN",
    "api_key": "your-porkbun-api-key",
    "secret_key": "your-porkbun-secret-key",
  }
}
```

## Metadata

This provider does not recognize any special metadata fields unique to Porkbun.

## Usage

An example `dnsconfig.js` configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_PORKBUN = NewDnsProvider("porkbun");

D("example.tld", REG_NONE, DnsProvider(DSP_PORKBUN),
    A("test", "1.2.3.4")
);
```
