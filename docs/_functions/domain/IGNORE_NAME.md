---
name: IGNORE_NAME
parameters:
  - pattern
---

WARNING: The `IGNORE_*` family  of functions is risky to use. The code
is brittle and has subtle bugs. Use at your own risk. Do not use these
commands with `D_EXTEND()`.

`IGNORE_NAME` can be used to ignore some records present in zone.
All records (independently of their type) of that name will be completely ignored.

`IGNORE_NAME` is like `NO_PURGE` except it acts only on some specific records instead of the whole zone.

Technically `IGNORE_NAME` is a promise that DNSControl will not add, change, or delete records at a given label.  This permits another entity to "own" that label.

`IGNORE_NAME` is generally used in very specific situations:

* Some records are managed by some other system and DNSControl is only used to manage some records and/or keep them updated. For example a DNS record that is managed by Kubernetes External DNS, but DNSControl is used to manage the rest of the zone. In this case we don't want DNSControl to try to delete the externally managed record.
* To work-around a pseudo record type that is not supported by DNSControl. For example some providers have a fake DNS record type called "URL" which creates a redirect. DNSControl normally deletes these records because it doesn't understand them. `IGNORE_NAME` will leave those records alone.

In this example, DNSControl will insert/update the "baz.example.com" record but will leave unchanged the "foo.example.com" and "bar.example.com" ones.

{% include startExample.html %}

```js
D("example.com",
  `IGNORE_NAME`("foo"),
  `IGNORE_NAME`("bar"),
  A("baz", "1.2.3.4")
);
```

{% include endExample.html %}

`IGNORE_NAME` also supports glob patterns in the style of the [gobwas/glob](https://github.com/gobwas/glob) library. All of
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

# Caveats

It is considered as an error to try to manage an ignored record.
Ignoring a label is a promise that DNSControl won't meddle with
anything at a particular label, therefore DNSControl prevents you from
adding records at a label that is `IGNORE_NAME`'ed.

Use `IGNORE_NAME("@")` to ignore at the domain's apex. Most providers
insert magic or unchangable records at the domain's apex; usually `NS`
and `SOA` records.  DNSControl treats them specially.

# Errors

* `trying to update/add IGNORE_NAME'd record: foo CNAME`

This means you have both ignored `foo` and included a record (in this
case, a CNAME) to update it.  This is an error because `IGNORE_NAME`
is a promise not to modify records at a certain label so that others
may have free reign there.  Therefore, DNSControl prevents you from
modifying that label.

The `foo CNAME` at the end of the message indicates the label name
(`foo`) and the type of record (`CNAME`) that your dnsconfig.js file
is trying to insert.

You can override this error by adding the
`IGNORE_NAME_DISABLE_SAFETY_CHECK` flag to the record.

    TXT('vpn', "this thing", IGNORE_NAME_DISABLE_SAFETY_CHECK)

Disabling this safety check creates two risks:

1. Two owners (DNSControl and some other entity) toggling a record between two settings.
2. The other owner wiping all records at this label, which won't be noticed until the next time dnscontrol is run.
