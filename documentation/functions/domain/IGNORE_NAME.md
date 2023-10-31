---
name: IGNORE_NAME
parameters:
  - pattern
  - rTypes
parameter_types:
  pattern: string
  rTypes: string?
---

`IGNORE_NAME(a)` is the same as `IGNORE(a, "*", "*")`.

`IGNORE_NAME(a, b)` is the same as `IGNORE(a, b, "*")`.
