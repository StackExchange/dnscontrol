---
name: IP
parameters:
  - ip
parameter_types:
  ip: string
return: number
---

Converts an IPv4 address from string to an integer. This allows performing mathematical operations with the IP address.

{% code title="dnsconfig.js" %}
```javascript
var addrA = IP("1.2.3.4")
var addrB = addrA + 1
// addrB = 1.2.3.5
```
{% endcode %}

{% hint style="info" %}
**NOTE**: `IP()` does not accept IPv6 addresses (PRs gladly accepted!). IPv6 addresses are simply strings:
{% endhint %}

{% code title="dnsconfig.js" %}
```javascript
// IPv4 Var
var addrA1 = IP("1.2.3.4");
var addrA2 = "1.2.3.4";

// IPv6 Var
var addrAAAA = "0:0:0:0:0:0:0:0";
```
{% endcode %}
