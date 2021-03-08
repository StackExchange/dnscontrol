---
layout: default
title: CAA Builder
---

# DMARC Builder

dnscontrol contains a DMARC_BUILDER which can be used to simply create
DMARC policies for your domains.


## Example

### Simple example

```js
DMARC_BUILDER({
  policy: 'reject',
  ruf: [
    'mailto:mailauth-reports@example.com',
  ],
})
```

This yield the following record:
```
@   IN  TXT "v=DMARC1; p=reject; ruf=mailto:mailauth-reports@example.com"
```

### Advanced example

```js
DMARC_BUILDER({
  policy: 'reject',
  subdomainPolicy: 'quarantine',
  percent: 50,
  alignmentSPF: 'r',
  alignmentDKIM: 'strict',
  rua: [
    'mailto:mailauth-reports@example.com',
    'https://dmarc.example.com/submit',
  ],
  ruf: [
    'mailto:mailauth-reports@example.com',
  ],
  failureOptions: '1',
  reportInterval: '1h',
}),

DMARC_BUILDER({
  label: 'insecure',
  policy: 'none',
  ruf: [
    'mailto:mailauth-reports@example.com',
  ],
  failureOptions: {
      SPF: false,
      DKIM: true,
  },
})
```

This yields the following records:

```
@           IN  TXT "v=DMARC1; p=reject; sp=quarantine; adkim=s; aspf=r; pct=50; rua=mailto:mailauth-reports@example.com,https://dmarc.example.com/submit; ruf=mailto:mailauth-reports@example.com; fo=1; ri=3600"
insecure    IN  TXT "v=DMARC1; p=none; ruf=mailto:mailauth-reports@example.com; fo=d"
```


### Parameters

* `label:` The DNS label for the DMARC record (`_dmarc` prefix is added, default: `'@'`)
* `policy:` The DMARC policy (`p=`), must be one of `'none'`, `'quarantine'`, `'reject'`
* `subdomainPolicy:` The DMARC policy for subdomains (`sp=`), must be one of `'none'`, `'quarantine'`, `'reject'` (optional)
* `alignmentSPF:` `'strict'`/`'s'` or `'relaxed'`/`'r'` alignment for SPF (`aspf=`, default: `'r'`)
* `alignmentDKIM:` `'strict'`/`'s'` or `'relaxed'`/`'r'` alignment for DKIM (`adkim=`, default: `'r'`)
* `percent:` Number between `0` and `100`, percentage for which policies are applied (`pct=`, default: `100`)
* `rua:` Array of aggregate report targets (optional)
* `ruf:` Array of failure report targets (optional)
* `failureOptions:` Object or string; Object containing booleans `SPF` and `DKIM`, string is passed raw (`fo=`, default: `'0'`)
* `failureFormat:` Format in which failure reports are requested (`rf=`, default: `'afrf'`)
* `reportInterval:` Interval in which reports are requested (`ri=`)
* `ttl:` Input for `TTL` method (optional)

### Caveats

* TXT records are automatically split using `AUTOSPLIT`.
* URIs in the `rua` and `ruf` arrays are passed raw. You must percent-encode all commas and exclamation points in the URI itself.
