---
name: IP
parameters:
  - ip
---

Converts the IP address from string to an integer. This allows performing mathematical operations with the IP address.

{% include startExample.html %}
{% highlight js %}

var addrA = IP('1.2.3.4')
var addrB = addrA + 1
// addrB = 1.2.3.5

{%endhighlight%}
{% include endExample.html %}
