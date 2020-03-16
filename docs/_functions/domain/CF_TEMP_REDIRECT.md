---
name: CF_TEMP_REDIRECT
parameters:
  - destination
  - modifiers...
---

`CF_TEMP_REDIRECT` uses Cloudflare-specific features ("Forwarding URL" Page
Rules) to generate a HTTP 302 temporary redirect.

If _any_ `CF_REDIRECT` or `CF_TEMP_REDIRECT` functions are used then
`dnscontrol` will manage _all_ "Forwarding URL" type Page Rules for the domain.
Page Rule types other than "Forwarding URL” will be left alone.

WARNING: Cloudflare does not currently fully document the Page Rules API and
this interface is not extensively tested. Take precautions such as making
backups and manually verifying `dnscontrol preview` output before running
`dnscontrol push`. This is especially true when mixing Page Rules that are
managed by DNSControl and those that aren't.

{% include startExample.html %}
{% highlight js %}
D("foo.com", .... ,
  CF_TEMP_REDIRECT("example.mydomain.com/*", "https://otherplace.yourdomain.com/$1"),
);
{%endhighlight%}
{% include endExample.html %}
