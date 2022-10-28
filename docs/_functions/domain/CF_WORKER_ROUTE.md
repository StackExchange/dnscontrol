---
name: CF_WORKER_ROUTE
parameters:
  - pattern
  - script
provider: CLOUDFLAREAPI
---

`CF_WORKER_ROUTE` uses the [Cloudflare Workers](https://developers.cloudflare.com/workers/)
API to manage [worker routes](https://developers.cloudflare.com/workers/platform/routes)
for a given domain.

If _any_ `CF_WORKER_ROUTE` function is used then `dnscontrol` will manage _all_
Worker Routes for the domain. To be clear: this means it will delete existing routes that
were created outside of DNSControl.

WARNING: This interface is not extensively tested. Take precautions such as making
backups and manually verifying `dnscontrol preview` output before running
`dnscontrol push`.

This example assigns the patterns `api.foo.com/*` and `foo.com/api/*` to a `my-worker` script:

{% capture example %}
```js
D("foo.com", .... ,
    CF_WORKER_ROUTE("api.foo.com/*", "my-worker"),
    CF_WORKER_ROUTE("foo.com/api/*", "my-worker"),
);
```
{% endcapture %}

{% include example.html content=example %}
