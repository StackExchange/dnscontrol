---
name: CF_WORKER_ROUTE
parameters:
  - pattern
  - script
---

`CF_WORKER_ROUTE` uses [Cloudflare Workers](https://developers.cloudflare.com/workers/) 
API to setup [routes](https://developers.cloudflare.com/workers/platform/routes)
for a given domain.

If _any_ `CF_WORKER_ROUTE` function is used then `dnscontrol` will manage _all_ 
Worker Routes for the domain.

WARNING: This interface is not extensively tested. Take precautions such as making
backups and manually verifying `dnscontrol preview` output before running
`dnscontrol push`. This is especially true when mixing Worker Routes that are
managed by DNSControl and those that aren't.

This example assigns the patterns `api.foo.com/*` and `foo.com/api/*` to a `my-worker` script:

{% include startExample.html %}
{% highlight js %}
D("foo.com", .... ,
    CF_WORKER_ROUTE("api.foo.com/*", "my-worker"),
    CF_WORKER_ROUTE("foo.com/api/*", "my-worker"),
);
{%endhighlight%}
{% include endExample.html %}
