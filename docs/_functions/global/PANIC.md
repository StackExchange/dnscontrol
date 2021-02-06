---
name: PANIC
parameters:
  - message
---

`PANIC` terminates the script and therefore DnsControl with an exit code of 1. This should be used if your script cannot gather enough information to generate records, for example when a HTTP request failed.

{% include startExample.html %}
{% highlight js %}
PANIC("Something really bad has happened");
{%endhighlight%}
{% include endExample.html %}
