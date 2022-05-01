---
name: Internet.bs
title: Internet.bs Provider
layout: default
jsId: INTERNETBS
---
# Internet.bs Provider

DNSControl's Internet.bs provider supports being a Registrar. Support for being a DNS Provider is not included, but could be added in the future.

## Configuration
In your credentials file, you must provide your API key and account password

```json
{
  "internetbs": {
    "TYPE": "INTERNETBS",
    "api-key": "your-api-key",
    "password": "account-password"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Internet.bs.

## Usage
Example Javascript:

```js
var REG_INTERNETBS = NewRegistrar('internetbs', 'INTERNETBS');

D("example.com", REG_INTERNETBS,
  NAMESERVER("ns1.example.com."),
  NAMESERVER("ns2.example.com."),
);
```

## Activation

Pay attention, you need to define white list of IP for API. But you always can change it on `My Profile > Reseller Settings`
