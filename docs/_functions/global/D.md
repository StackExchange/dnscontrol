---
name: D
parameters:
  - name
  - registrar
  - modifiers...
---

`D` adds a new Domain for DNSControl to manage. The first two arguments are required: the domain name (fully qualified `example.com` without a trailing dot), and the
name of the registrar (as previously declared with [NewRegistrar](#NewRegistrar)). Any number of additional arguments may be included to add DNS Providers with [DNSProvider](#DNSProvider),
add records with [A](#A), [CNAME](#CNAME), and so forth, or add metadata.

Modifier arguments are processed according to type as follows:

- A function argument will be called with the domain object as it's only argument. Most of the [built-in modifier functions](#domain-modifiers) return such functions.
- An object argument will be merged into the domain's metadata collection.
- An array argument will have all of it's members evaluated recursively. This allows you to combine multiple common records or modifiers into a variable that can
   be used like a macro in multiple domains.

{% include startExample.html %}
{% highlight js %}
var REGISTRAR = NewRegistrar("name.com", "NAMEDOTCOM");
var r53 = NewDnsProvider("R53","ROUTE53");

// simple domain
D("example.com", REGISTRAR, DnsProvider(r53),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com.")
);

// "macro" for records that can be mixed into any zone
var GOOGLE_APPS_DOMAIN_MX = [
    MX('@', 1, 'aspmx.l.google.com.'),
    MX('@', 5, 'alt1.aspmx.l.google.com.'),
    MX('@', 5, 'alt2.aspmx.l.google.com.'),
    MX('@', 10, 'alt3.aspmx.l.google.com.'),
    MX('@', 10, 'alt4.aspmx.l.google.com.'),
]

D("example.com", REGISTRAR, DnsProvider(r53),
  A("@","1.2.3.4"),
  CNAME("test", "foo.example2.com."),
  GOOGLE_APPS_DOMAIN_MX
);
{%endhighlight%}
{% include endExample.html %}


# Split Horizon DNS

DNSControl supports Split Horizon DNS. Simply
define the domain two or more times, each with
their own unique parameters.

To differentiate the different domains, specify the domains as
1domain.tld!tag`, such as `example.com!inside` and
`example.com!outside`.

{% include startExample.html %}
{% highlight js %}
var REG = NewRegistrar("Third-Party", "NONE");
var DNS_INSIDE = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");
var DNS_OUTSIDE = NewDnsProvider("bind", "BIND");

D("example.com!inside", REG, DnsProvider(DNS_INSIDE),
  A("main", "1.1.1.1")
);

D("example.com!outside", REG, DnsProvider(DNS_OUTSIDE),
  A("main", "8.8.8.8")
);

D_EXTEND("example.com!inside",
  A("main", "11.11.11.11")
);
{%endhighlight%}
{% include endExample.html %}

A domain name without a `!` defaults to has a tag that is the empty
string. For example, `example.com` and `example.com!` are equivalent.
However, we strongly recommend that all horizons have tags to prevent
human mistakes. In other words, if you have `domain.tld` and
`domain.tld!external` you now require humans to remember that
`domain.tld` is the external one.  I mean... the internal one.  You
may have noticed this mistake, but will you in 6 months?  Will your
corworkers You
get the idea.

DNSControl command line flag `--domains` is an exact match. That is,
if you define domains `example.com!one` and `example.com!two`, the
flag `--domains=example.com` will match zero domains.
`--domains='example.com!two` will match the last one.  The quotes are
required if your shell treats `!` as a special character, which is
probably does.  If you see an error that mentions `event not found`
you probably forgot the quotes.
