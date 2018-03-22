---
name: CF_REDIRECT
parameters:
  - destination
  - modifiers...
---

`CF_REDIRECT` is the same as `CF_TEMP_REDIRECT` but generates a
http 301 redirect (permanent redirect) instead of a temporary
redirect.

These redirects are cached by browsers forever, usually ignoring
any TTLs or other cache invalidation techniques.   It should be
used with great care.  We suggest using a `CF_TEMP_REDIRECT`
initially, then changing to a `CF_REDIRECT` only after sufficient
time has elapsed to prove this is what you really want.
