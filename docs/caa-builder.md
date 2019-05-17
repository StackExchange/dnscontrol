---
layout: default
title: CAA Builder
---

# CAA Builder

dnscontrol contains a CAA_BUILDER which can be used to simply create
CAA records for your domains. Instead of creating each CAA record
individually, you can simply configure your report mail address, the
authorized certificate authorities and the builder cares about the rest.


## Example

For example you can use:

```
CAA_BUILDER({
  label: "@",
  iodef: "test@domain.tld",
  iodef_critical: true,
  issue: [
    "letsencrypt.org",
    "comodoca.com",
  ],
  issuewild: [
    ";",
  ]
})
```

The parameters are:

* `label:` The label of the CAA record. (Optional. Default: `"@"`)
* `iodef:` Report all violation to configured mail address.
* `iodef_critical:` This can be `true` or `false`. If enabled and CA does not support this record, then certificate issue will be refused. (Optional. Default: `false`)
* `issue:` An array of CAs which are allowed to issue certificates.
* `issuewild:` An array of CAs which are allowed to issue wildcard certificates. (Can be simply `;` for forbid issuing wildcard certificates for all CAs)

`CAA_BUILDER()` returns multiple `TXT()` records (when configured as example above):

  * `CAA("@", "iodef", "test@domain.tld", CAA_CRITICAL)`
  * `CAA("@", "issue", "letsencrypt.org")`
  * `CAA("@", "issue", "comodoca.com")`
  * `CAA("@", "issuewild", ";")`

```
