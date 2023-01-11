---
name: R53_ZONE
parameters:
  - zone_id
provider: ROUTE53
---

R53_ZONE lets you specify the AWS Zone ID for an entire domain (D()) or a specific R53_ALIAS() record.

When used with D(), it sets the zone id of the domain. This can be used to differentiate between split horizon domains in public and private zones.

When used with R53_ALIAS() it sets the required Route53 hosted zone id in a R53_ALIAS record. See [R53_ALIAS's documentation](https://stackexchange.github.io/dnscontrol/js#R53_ALIAS) for details.
