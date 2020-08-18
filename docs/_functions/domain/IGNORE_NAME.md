---
name: IGNORE_NAME
---

IGNORE_NAME can be used to ignore some records present in zone.
All records (independently of their type) of that name will be completely ignored.

IGNORE_NAME is like NO_PURGE except it acts only on some specific records instead of the whole zone.

IGNORE_NAME is generally used in very specific situations:

* Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS record that is managed by Kubernetes External DNS, but DNSControl is used to manage the rest of the zone. In this case we don't want DNSControl to try to delete the externally managed record.
* To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. IGNORE_NAME will leave those records alone.

In this example, DNSControl will insert/update the "baz.example.com" record but will leave unchanged the "foo.example.com" and "bar.example.com" ones.

{% include startExample.html %}
{% highlight js %}
D("example.com",
  IGNORE_NAME("foo"),
  IGNORE_NAME("bar"),
  A("baz", "1.2.3.4")
);
{%endhighlight%}
{% include endExample.html %}

IGNORE_NAME also supports glob patterns in the style of the [gobwas/glob](https://github.com/gobwas/glob) library. All of
the following patterns will work:

* `IGNORE_NAME("*.foo")` will ignore all records in the style of `bar.foo`, but will not ignore records using a double
subdomain, such as `foo.bar.foo`.
* `IGNORE_NAME("**.foo")` will ignore all subdomains of `foo`, including double subdomains.
* `IGNORE_NAME("?oo")` will ignore all records of three symbols ending in `oo`, for example `foo` and `zoo`. It will
not match `.`
* `IGNORE_NAME("[abc]oo")` will ignore records `aoo`, `boo` and `coo`. `IGNORE_NAME("[a-c]oo")` is equivalent.
* `IGNORE_NAME("[!abc]oo")` will ignore all three symbol records ending in `oo`, except for `aoo`, `boo`, `coo`. `IGNORE_NAME("[!a-c]oo")` is equivalent.
* `IGNORE_NAME("{bar,[fz]oo}")` will ignore `bar`, `foo` and `zoo`.
* `IGNORE_NAME("\\*.foo")` will ignore the literal record `*.foo`.

It is considered as an error to try to manage an ignored record.