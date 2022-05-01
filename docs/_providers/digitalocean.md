---
name: DigitalOcean
title: DigitalOcean Provider
layout: default
jsId: DIGITALOCEAN
---
# DigitalOcean Provider

## Configuration
In your credentials file, you must provide your
[DigitalOcean OAuth Token](https://cloud.digitalocean.com/settings/applications)

```json
{
  "digitalocean": {
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
var REG_NONE = NewRegistrar('none', 'NONE')
var DIGITALOCEAN = NewDnsProvider("digitalocean", "DIGITALOCEAN");

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
