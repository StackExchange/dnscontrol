---
name: CF_TEMP_REDIRECT
parameters:
  - destination
  - modifiers...
---

`CF_REDIRECT` uses CloudFlare-specific features ("page rules") to
generate an HTTP 301 redirect.

WARNING: If the domain has other pagerules in place, they may be
deleted. At this time this feature is best used on bare domains
that need to redirect to another domain, perhaps with wildcard
substitutions.

{% include startExample.html %}
{% highlight js %}
D("foo.com", .... ,
  CF_TEMP_REDIRECT("example.mydomain.com/*", "https://otherplace.yourdomain.com/$1"),
);
{%endhighlight%}
{% include endExample.html %}
