---
name: Linode
title: Linode Provider
layout: default
jsId: LINODE
---
# Linode Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `LINODE`
along with your [Linode Personal Access Token](https://cloud.linode.com/profile/tokens).

Example:

```json
{
  "linode": {
    "TYPE": "LINODE",
    "token": "your-linode-personal-access-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Linode.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("none");
var DSP_LINODE = NewDnsProvider("linode");

D("example.tld", REG_NONE, DnsProvider(DSP_LINODE),
    A("test", "1.2.3.4")
);
```

## Activation
[Create Personal Access Token](https://cloud.linode.com/profile/tokens)

## Caveats
Linode does not allow all TTLs, but only a specific subset of TTLs. The following TTLs are supported
([source](https://github.com/linode/manager/blob/master/src/domains/components/SelectDNSSeconds.js)):

- 300
- 3600
- 7200
- 14400
- 28800
- 57600
- 86400
- 172800
- 345600
- 604800
- 1209600
- 2419200

The provider will automatically round up your TTL to one of these values. For example, 600 seconds would become 3600
seconds, but 300 seconds would stay 300 seconds.
