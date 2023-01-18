---
name: AKAMAICDN
parameters:
  - name
  - target
  - modifiers...
provider: AKAMAIEDGEDNS
parameter_types:
  name: string
  target: string
  "modifiers...": RecordModifier[]
---

AKAMAICDN is a proprietary record type that is used to configure [Zone Apex Mapping](https://blogs.akamai.com/2019/08/fast-dns-zone-apex-mapping-dnssec.html).
The AKAMAICDN target must be preconfigured in the Akamai network.
