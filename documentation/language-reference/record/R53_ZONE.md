---
name: R53_ZONE
parameters:
  - zone_id
parameter_types:
  zone_id: string
ts_return: DomainModifier & RecordModifier
provider: ROUTE53
---

`R53_ZONE` lets you specify the AWS Zone ID for an entire domain ([`D()`](../top-level-functions/D.md)) or a specific [`R53_ALIAS()`](../domain/R53_ALIAS.md) record.

When used with [`D()`](../top-level-functions/D.md), it sets the zone id of the domain. This can be used to differentiate between split horizon domains in public and private zones. See this [example](../../providers/route53.md#split-horizon) in the [Amazon Route 53 provider page](../../providers/route53.md).

When used with [`R53_ALIAS()`](../domain/R53_ALIAS.md) it sets the required Route53 hosted zone id in a R53_ALIAS record. See [`R53_ALIAS()`](../domain/R53_ALIAS.md) documentation for details.
