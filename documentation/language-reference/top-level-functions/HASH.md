---
name: HASH
parameters:
  - algorithm
  - value
parameter_types:
  algorithm: '"SHA1" | "SHA256" | "SHA512"'
  value: string
ts_return: string
---

`HASH` hashes `value` using the hashing algorithm given in `algorithm`
(accepted values `SHA1`, `SHA256`, and `SHA512`) and returns the hex encoded
hash value.

example `HASH("SHA1", "abc")` returns `a9993e364706816aba3e25717850c26c9cd0d89d`.

`HASH()`'s primary use case is for managing [catalog zones](https://datatracker.ietf.org/doc/html/rfc9432):

> a method for automatic DNS zone provisioning among DNS primary and secondary name
> servers by storing and transferring the catalog of zones to be provisioned as one
> or more regular DNS zones.

Here's an example of a catalog zone:

{% code title="dnsconfig.js" %}
```javascript
foo_name_suffix = HASH("SHA1", "foo.name") + ".zones"
D("catalog.example"
    [...]
    , TXT("version", "2")
    , PTR(foo_name_suffix, "foo.name.")
    , A("primaries.ext." + foo_name_suffix, "192.168.1.1")
)
```
{% endcode %}
