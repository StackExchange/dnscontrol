---
name: DigitalOcean
title: DigitalOcean Provider
layout: default
jsId: DIGITALOCEAN
---
# DigitalOcean Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DIGITALOCEAN`
along with your [DigitalOcean OAuth Token](https://cloud.digitalocean.com/settings/applications).

Example:

```json
{
  "mydigitalocean": {
    "TYPE": "DIGITALOCEAN",
    "token": "your-digitalocean-ouath-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to DigitalOcean.

## Usage
Example Javascript:

```js
var REG_NONE = NewRegistrar('none');
var DIGITALOCEAN = NewDnsProvider("mydigitalocean");

D("example.tld", REG_NONE, DnsProvider(DIGITALOCEAN),
    A("test","1.2.3.4")
);
```

## Activation
[Create OAuth Token](https://cloud.digitalocean.com/settings/applications)

## Limitations

- Digitalocean DNS doesn't support `;` value with CAA-records ([DigitalOcean documentation](https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records/))
- While Digitalocean DNS supports TXT records with multiple strings,
  their length is limited by the max API request of 512 octets.
