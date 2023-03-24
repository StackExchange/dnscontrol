---
name: R53_ZONE
parameters:
  - zone_id
parameter_types:
  zone_id: string
ts_return: DomainModifier & RecordModifier
provider: ROUTE53
---

R53_ZONE lets you specify the AWS Zone ID for an entire domain (D()) or a specific R53_ALIAS() record.

When used with D(), it sets the zone id of the domain. This can be used to differentiate between split horizon domains in public and private zones.

When used with R53_ALIAS() it sets the required Route53 hosted zone id in a R53_ALIAS record. See [R53_ALIAS's documentation](../../../domain_modifier_functions/service_provider_specific/amazon_route_53/R53_ALIAS.md) for details.
