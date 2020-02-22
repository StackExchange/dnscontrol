---
name: AUTODNSSEC
---

AUTODNSSEC indicates that the DNS provider can automatically manage
DNSSEC for a domain and we should ask it to do so.

At this time, AUTODNSSEC takes no parameters.
There is no ability to tune what the DNS provider sets, no algorithm choice.
We simply ask that they follow their defaults when enabling a no-fuss DNSSEC
data model.

{% include startExample.html %}
{% highlight js %}
D("example.com", .... ,
  AUTODNSSEC,
);
{%endhighlight%}
{% include endExample.html %}
