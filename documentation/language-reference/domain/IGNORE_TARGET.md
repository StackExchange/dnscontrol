---
name: IGNORE_TARGET
parameters:
  - pattern
  - rType
parameter_types:
  pattern: string
  rType: string
---

`IGNORE_TARGET_NAME(target)` is the same as `IGNORE("*", "*", target)`.

`IGNORE_TARGET_NAME(target, rtype)` is the same as `IGNORE("*", rtype, target)`.
