---
name: IGNORE
---

IGNORE can be used to ignore some records presents in zone.
All records (independently of their type) of that name will be completely ignored.

IGNORE is like NO_PURGE except it acts only on some specific records intead of the whole zone.

IGNORE is generally used in very specific situations:

* Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS record that is managed by Kubernetes External DNS, but DNSControl is used to manage the rest of the zone. In this case we don't want dnscontrol to try to delete the externally managed record.
* To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. IGNORE will leave those records alone.

In this example, dnscontrol will insert/update the "baz.example.com" record but will leave unchanged the "foo.example.com" and "bar.example.com" ones.

{% include startExample.html %}
{% highlight js %}
D("example.com",
  IGNORE("foo"),
  IGNORE("bar"),
  A("baz", "1.2.3.4")
);
{%endhighlight%}
{% include endExample.html %}

It is considered as an error to try to manage an ignored record.