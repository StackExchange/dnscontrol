---
name: IGNORE
ts_ignore: true
---

`IGNORE()` makes it possible for DNSControl to share management of a domain with an
external system.  The parameters of `IGNORE()` indicate which records are managed
elsewhere and should not be touched.

Suppose a domain is managed by both DNSControl and a third-party system. This creates
a problem because DNSControl will try to delete records inserted by the other system.  The
other system may get confused and re-insert those records.  The two systems will always
be modifying the records.

To solve this problem simply include `IGNORE()` statements that identify which records
are managed elsewhere and should be ignored

Technically `IGNORE_NAME` is a promise that DNSControl will not modify or
delete existing records that match particular patterns. It is like
[`NO_PURGE`](../domain/NO_PURGE.md) that matches only specific records.

Including a record that is ignored is considered an error and may have undefined behavior.

## Syntax

```
IGNORE(labelSpec, typeSpec, targetSpec):
IGNORE(labelSpec, typeSpec):
IGNORE(labelSpec):
```

* `labelSpec` is a glob that matches the DNS label. For example `"foo"` or `"foo*"`.  `"*"` matches all labels, as does the empty string (`""`).
* `typeSpec` is a comma-separated list of DNS types.  `"*"` matches all DNS types, as does the empty string (`""`).  For example "A" matches DNS A records, "A,CNAME" matches both A and CNAME records.
* `targetSpec` is a glob that matches the DNS target. For example `"foo"` or `"foo*"`.  `"*"` matches all targets, as does the empty string (`""`).

typeSpec and targetSpec default to `"*"` if they are omitted.

## Globs

The `labelSpec` and `targetSpec` parameters supports glob patterns in the style
of the [gobwas/glob](https://github.com/gobwas/glob) library.  All of the
following patterns will work:

* `IGNORE("*.foo")` will ignore all records in the style of `bar.foo`, but will not ignore records using a double
subdomain, such as `foo.bar.foo`.
* `IGNORE("**.foo")` will ignore all subdomains of `foo`, including double subdomains.
* `IGNORE("?oo")` will ignore all records of three symbols ending in `oo`, for example `foo` and `zoo`. It will
not match `.`
* `IGNORE("[abc]oo")` will ignore records `aoo`, `boo` and `coo`. `IGNORE("[a-c]oo")` is equivalent.
* `IGNORE("[!abc]oo")` will ignore all three symbol records ending in `oo`, except for `aoo`, `boo`, `coo`.        `IGNORE("[!a-c]oo")` is equivalent.
* `IGNORE("{bar,[fz]oo}")` will ignore `bar`, `foo` and `zoo`.
* `IGNORE("\\*.foo")` will ignore the literal record `*.foo`.

## Examples

General examples:

{% code title="dnsconfig.js" %}
```javascript
D("example.com",
  IGNORE("foo"), // matches any records on foo.example.com
  IGNORE("baz", "A"), // matches any A records on label baz.example.com
  IGNORE("*", "MX", "*"), // matches all MX records
  IGNORE("*", "CNAME", "dev-*"), // matches CNAMEs pointing to hosts that start with dev-*
  IGNORE("bar", "A,MX"), // ignore only A and MX records for name bar
  IGNORE("*", "*", "dev-*), // Ignore targets with a `dev-` prefix
  IGNORE("*", "A", "1\.2\.3\."), // Ignore targets that match the 1.2.3.0/24 CIDR block.
);
```
{% endcode %}

Ignore Let's Encrypt (ACME) validation records:

```
  IGNORE("_acme-challenge.**", "TXT"),
```

Ignore DNS records typically inserted by Microsoft ActiveDirectory:

```
  IGNORE("_gc.**", "SRV"), // General Catalog
  IGNORE("_kerberos.**", "SRV"), // Kerb5 server
  IGNORE("_kpasswd.**", "SRV"), // Kpassword
  IGNORE("_ldap.**", "SRV"), // LDAP
  IGNORE("_msdcs", "NS"), // Microsoft Domain Controller Service
  IGNORE("_vlmcs.**", "SRV"), // FQDN of the KMS host
  IGNORE("domaindnszones", "A"),
  IGNORE("forestdnszones", "A"),
```

## Don't insert and ignore the same item

It is considered as an error to try to ignore records that
you yourself are inserting into a domain.

This will generate an error:

```
D("example.com", ...
    ...
    TXT("myhost", "mytext"),
    IGNORE("myhost", "*", "*"),
    ...
```

To disable this safety check, add the `DISABLE_IGNORE_SAFETY_CHECK` statement to the `D()`.

```
D("example.com", ...
    DISABLE_IGNORE_SAFETY_CHECK,
    ...
    TXT("myhost", "mytext"),
    IGNORE("myhost", "*", "*"),
    ...
```

FYI: Previously DNSControl permitted disabling this check on
a per-record basis using `IGNORE_NAME_DISABLE_SAFETY_CHECK`:

```
    // THIS NO LONGER WORKS! Use DISABLE_IGNORE_SAFETY_CHECK instead.
    TXT("myhost", "mytext", IGNORE_NAME_DISABLE_SAFETY_CHECK),
```

The `IGNORE_NAME_DISABLE_SAFETY_CHECK` feature does not exist in the diff2 world and its use will result in a validation error.

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
2. The other owner wiping all records at this label, which won't be noticed until the next time DNSControl is run.



## Caveats

{% hint style="warning" %}
**WARNING**: The `IGNORE_*` family  of functions is complex and complex things are risky. Test extensively.
{% endhint %}

* `IGNORE` is not tested with `D_EXTEND()` and may not work.
* There is no locking.  If the external system and DNSControl make updates at the exact same time, the results are undefined.
* `targetSpec` does not match fields other than the primary target.  For example, `MX` records have a target hostname plus a priority. There is no way to match the priority.
* The BIND provider can not ignore records it doesn't know about.  If it does not have access to an existing zonefile, it will create a zonefile from scratch. That new zonefile will not have any external records.

