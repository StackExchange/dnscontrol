---
name: D
parameters:
  - name
  - registrar
  - modifiers...
parameter_types:
  name: string
  registrar: string
  "modifiers...": DomainModifier[]
---

`D` adds a new Domain for DNSControl to manage. The first two arguments are required: the domain name (fully qualified `example.com` without a trailing dot), and the
name of the registrar (as previously declared with [NewRegistrar](NewRegistrar.md)). Any number of additional arguments may be included to add DNS Providers with [DNSProvider](NewDnsProvider.md),
add records with [A](../domain-modifiers/A.md), [CNAME](../domain-modifiers/CNAME.md), and so forth, or add metadata.

Modifier arguments are processed according to type as follows:

- A function argument will be called with the domain object as it's only argument. Most of the [built-in modifier functions](https://docs.dnscontrol.org/language-reference/domain-modifiers-modifiers) return such functions.
- An object argument will be merged into the domain's metadata collection.
- An array argument will have all of it's members evaluated recursively. This allows you to combine multiple common records or modifiers into a variable that can
   be used like a macro in multiple domains.

{% code title="dnsconfig.js" %}
```javascript
// simple domain
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com.")
);

// "macro" for records that can be mixed into any zone
var GOOGLE_APPS_DOMAIN_MX = [
    MX("@", 1, "aspmx.l.google.com."),
    MX("@", 5, "alt1.aspmx.l.google.com."),
    MX("@", 5, "alt2.aspmx.l.google.com."),
    MX("@", 10, "alt3.aspmx.l.google.com."),
    MX("@", 10, "alt4.aspmx.l.google.com."),
]

D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com."),
  GOOGLE_APPS_DOMAIN_MX
);
```
{% endcode %}


# Split Horizon DNS

DNSControl supports Split Horizon DNS. Simply
define the domain two or more times, each with
their own unique parameters.

To differentiate the different domains, specify the domains as
`domain.tld!tag`, such as `example.com!inside` and
`example.com!outside`.

{% code title="dnsconfig.js" %}
```javascript
var REG_THIRDPARTY = NewRegistrar("ThirdParty");
var DNS_INSIDE = NewDnsProvider("Cloudflare");
var DNS_OUTSIDE = NewDnsProvider("bind");

D("example.com!inside", REG_THIRDPARTY, DnsProvider(DNS_INSIDE),
  A("www", "10.10.10.10")
);

D("example.com!outside", REG_THIRDPARTY, DnsProvider(DNS_OUTSIDE),
  A("www", "20.20.20.20")
);

D_EXTEND("example.com!inside",
  A("internal", "10.99.99.99")
);
```
{% endcode %}

A domain name without a `!` is assigned a tag that is the empty
string. For example, `example.com` and `example.com!` are equivalent.
However, we strongly recommend against using the empty tag, as it
risks creating confusion.  In other words, if you have `domain.tld`
and `domain.tld!external` you now require humans to remember that
`domain.tld` is the external one.  I mean... the internal one.  You
may have noticed this mistake, but will your coworkers?  Will you in
six months? You get the idea.

DNSControl command line flag `--domains` matches the full name (with the "!").  If you
define domains `example.com!george` and `example.com!john` then:

* `--domains=example.com` will not match either domain.
* `--domains='example.com!george'` will match only match the first.
* `--domains='example.com!george",example.com!john` will match both.

{% hint style="info" %}
**NOTE**: The quotes are required if your shell treats `!` as a special
character, which is probably does.  If you see an error that mentions
`event not found` you probably forgot the quotes.
{% endhint %}
