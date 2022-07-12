---
name: DNSimple
title: DNSimple Provider
layout: default
jsId: DNSIMPLE
---
# DNSimple Provider
## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DIGITALOCEAN`
along with a DNSimple account access token.

Example:

```json
{
  "dnsimple": {
    "TYPE": "DNSIMPLE",
    "token": "your-dnsimple-account-access-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to DNSimple.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_DNSIMPLE = NewRegistrar("dnsimple");
var DSP_DNSIMPLE = NewDnsProvider("dnsimple");

D("example.tld", REG_DNSIMPLE, DnsProvider(DSP_DNSIMPLE),
    A("test", "1.2.3.4")
);
```

## Activation
DNSControl depends on a DNSimple account access token.

## Caveats

### CAA

As of July 2022, the DNSimple DNS does not accept spaces in CAA records. Putting spaces in the record will result in a 400 Validation Failed error.

```
0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

Removing the spaces will work.
```
0 issue "letsencrypt.org;validationmethods=dns-01;accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

