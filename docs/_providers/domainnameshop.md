---
name: Domainnameshop
title: Domainnameshop Provider
jsId: DOMAINNAMESHOP
---
# DOMAINNAMESHOP Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DOMAINNAMESHOP`
along with your [Domainnameshop Token and Secret](https://www.domeneshop.no/admin?view=api).

Example:

```json
{
  "mydomainnameshop": {
    "TYPE": "DOMAINNAMESHOP",
    "token": "your-domainnameshop-token",
    "secret": "your-domainnameshop-secret"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Domainnameshop.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("none");
var DSP_DOMAINNAMESHOP = NewDnsProvider("mydomainnameshop");

D("example.tld", REG_NONE, DnsProvider(DSP_DOMAINNAMESHOP),
    A("test", "1.2.3.4")
);
```

## Activation
[Create API Token and secret](https://www.domeneshop.no/admin?view=api)

## Limitations

- Domainnameshop DNS only supports TTLs which are a multiple of 60.