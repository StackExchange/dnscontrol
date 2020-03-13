---
name: CF_REDIRECT
parameters:
  - destination
  - modifiers...
---

`CF_REDIRECT` uses Cloudflare-specific features ("Forwarding URL"
Page Rules) to generate a HTTP 301 permanent redirect.

If _any_ `CF_REDIRECT` or `CF_TEMP_REDIRECT` functions are used
then `dnscontrol` will manage _all_ "Forwarding URL" type Page Rules
for the domain. Page Rule types other than "Forwarding URL” will
be left alone.

These redirects are cached by browsers forever, usually ignoring
any TTLs or other cache invalidation techniques.   It should be
used with great care.  We suggest using a `CF_TEMP_REDIRECT`
initially, then changing to a `CF_REDIRECT` only after sufficient
time has elapsed to prove this is what you really want.

This example redirects the bare (aka apex, or naked) domain to www:

{% include startExample.html %}
{% highlight js %}
D("foo.com", .... ,
  CF_REDIRECT("mydomain.com/*", "https://www.mydomain.com/$1"),
);
{%endhighlight%}
{% include endExample.html %}
