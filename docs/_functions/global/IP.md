---
name: IP
parameters:
  - ip
---

Converts an IPv4 address from string to an integer. This allows performing mathematical operations with the IP address.

This does not accept IPv6 addresses. (PRs gladly accepted.)

{% include startExample.html %}
{% highlight js %}

var addrA = IP('1.2.3.4')
var addrB = addrA + 1
// addrB = 1.2.3.5

{%endhighlight%}
{% include endExample.html %}
