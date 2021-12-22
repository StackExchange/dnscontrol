---
name: R53_ZONE
parameters:
  - zone_id
---

R53_ZONE lets you specify the AWS Zone ID for an entire domain (D()) or a specific R53_ALIAS() record.

When used with D(), it sets the zone id of the domain. This can be used to differentiate between split horizon domains in public and private zones.

When used with R53_ALIAS() it sets the required Route53 hosted zone id in a R53_ALIAS record. See [https://stackexchange.github.io/dnscontrol/js#R53_ALIAS](R53_ALIAS's documentation) for details.




