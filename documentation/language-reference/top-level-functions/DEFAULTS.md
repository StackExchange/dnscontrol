---
name: DEFAULTS
parameters:
  - modifiers...
parameter_types:
  "modifiers...": DomainModifier[]
---

`DEFAULTS` allows you to declare a set of default arguments to apply to all subsequent domains. Subsequent calls to [`D`](D.md) will have these
arguments passed as if they were the first modifiers in the argument list.

## Example

We want to create backup zone files for all domains, but not actually register them. Also create a [`DefaultTTL`](../domain-modifiers/DefaultTTL.md).
The domain `example.com` will have the defaults set.

{% code title="dnsconfig.js" %}
```javascript
var COMMON = NewDnsProvider("foo");
DEFAULTS(
  DnsProvider(COMMON, 0),
  DefaultTTL("1d")
);

D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("@","1.2.3.4")
);
```
{% endcode %}

If you want to clear the defaults, you can do the following.
The domain `example2.com` will **not** have the defaults set.

{% code title="dnsconfig.js" %}
```javascript
DEFAULTS();

D("example2.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("@","1.2.3.4")
);
```
{% endcode %}
