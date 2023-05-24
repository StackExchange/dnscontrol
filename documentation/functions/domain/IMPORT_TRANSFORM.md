---
name: IMPORT_TRANSFORM
parameters:
  - transform table
  - domain
  - modifiers...
ts_ignore: true
---

{% hint style="warning" %}
Don't use this feature. It was added for a very specific situation at Stack Overflow.
{% endhint %}

`IMPORT_TRANSFORM` adds to the domain all the records from another
domain, after making certain transformations and resetting the TTL.

Example:

Suppose foo.com is a regular domain.  bar.com is a regular domain,
but certain records should be the same as foo.com with these
exceptions: "bar.com" is added to the name, the TTL is changed to
300, if the IP address is between 1.2.3.10 and 1.2.3.20 then rewrite
the IP address to be based on 123.123.123.100 (i.e. .113 or .114).

You wouldn't want to maintain bar.com manually, would you?  It would
be very error prone. Therefore instead you maintain foo.com and
let `IMPORT_TRANSFORM` automatically generate bar.com.

```text
foo.com:
    one.foo.com.    IN A 1.2.3.1
    two.foo.com.    IN A 1.2.3.2
    three.foo.com.  IN A 1.2.3.13
    four.foo.com.   IN A 1.2.3.14

bar.com:
    www.bar.com.    IN 123.123.123.123
    one.foo.com.bar.com.    IN A 1.2.3.1
    two.foo.com.bar.com.    IN A 1.2.3.2
    three.foo.com.bar.com.  IN A 123.123.123.113
    four.foo.com.bar.com.   IN A 123.123.123.114
```

Here's how you'd implement this in DNSControl:

{% code title="dnsconfig.js" %}
```javascript
var TRANSFORM_INT = [
    // RANGE_START, RANGE_END, NEW_BASE
    { low: "1.2.3.10", high: "1.2.3.20", newBase: "123.123.123.100" },  //   .10 to .20 rewritten as 123.123.123.100+IP
    { low: "2.4.6.80", high: "2.4.6.90", newBase: "123.123.123.200" },  //   Another rule, just to show that you can have many.
]

D("foo.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("one","1.2.3.1")
  A("two","1.2.3.2")
  A("three","1.2.3.13")
  A("four","1.2.3.14")
);

D("bar.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("www","123.123.123.123")
  IMPORT_TRANSFORM(TRANSFORM_INT, "foo.com", 300),
);
```
{% endcode %}

Transform rules are: RANGE_START, RANGE_END, NEW_BASE.  NEW_BASE may be:

* An IP address.  Rebase the IP address on this IP address. Extract the host part of the /24 and add it to the "new base" address.
* A list of IP addresses. For each A record, inject an A record for each item in the list: `newBase: ["1.2.3.100", "2.4.6.8.100"]` would produce 2 records for each A record.
