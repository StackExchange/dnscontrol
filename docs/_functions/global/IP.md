---
name: IP
parameters:
  - ip
parameter_types:
  ip: string
return: number
---

Converts an IPv4 address from string to an integer. This allows performing mathematical operations with the IP address.

{% capture example %}
```js
var addrA = IP('1.2.3.4')
var addrB = addrA + 1
// addrB = 1.2.3.5
```
{% endcapture %}

{% include example.html content=example %}

NOTE: `IP()` does not accept IPv6 addresses (PRs gladly accepted!). IPv6 addresses are simply strings:

{% capture example2 %}
```js
// IPv4 Var
var addrA1 = IP("1.2.3.4");
var addrA2 = "1.2.3.4";

// IPv6 Var
var addrAAAA = "0:0:0:0:0:0:0:0";
```
{% endcapture %}

{% include example.html content=example2 %}
