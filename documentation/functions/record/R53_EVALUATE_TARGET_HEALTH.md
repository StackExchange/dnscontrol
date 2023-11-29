---
name: R53_EVALUATE_TARGET_HEALTH
parameters:
  - enabled
parameter_types:
  enabled: boolean
ts_return: RecordModifier
provider: ROUTE53
---

`R53_EVALUATE_TARGET_HEALTH` lets you enable target health evaluation for a [`R53_ALIAS()`](../domain/R53_ALIAS.md) record. Omitting `R53_EVALUATE_TARGET_HEALTH()` from `R53_ALIAS()` set the behavior to false.
