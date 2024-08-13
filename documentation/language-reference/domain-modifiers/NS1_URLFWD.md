---
name: NS1_URLFWD
parameters:
  - name
  - target
  - modifiers...
provider: NS1
parameter_types:
  name: string
  target: string
  "modifiers...": RecordModifier[]
---

`NS1_URLFWD` is an NS1-specific feature that maps to NS1's URLFWD record, which creates HTTP 301 (permanent) or 302 (temporary) redirects.


{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  NS1_URLFWD("urlfwd", "/ http://example.com 302 2 0")
);
```
{% endcode %}

The fields are:
* name: the record name
* target: a complex field containing the following, space separated:
    * from - the path to match
    * to - the url to redirect to
    * redirectType - (0 - masking, 301, 302)
    * pathForwardingMode - (0 - All, 1 - Capture, 2 - None)
    * queryForwardingMode - (0 - disabled, 1 - enabled)

{% hint style="warning" %}
**WARNING**: According to NS1, this type of record is deprecated and in the process
of being replaced by the premium-only `REDIRECT` record type. While still able to be
configured through the API, as suggested by NS1, please try not to use it, going forward.
{% endhint %}
