---
name: CF_SINGLE_REDIRECT
parameters:
  - name
  - code
  - when
  - then
  - modifiers...
provider: CLOUDFLAREAPI
parameter_types:
  name: string
  code: number
  when: string
  then: string
  "modifiers...": RecordModifier[]
---

`CF_SINGLE_REDIRECT` is a Cloudflare-specific feature for creating HTTP 301
(permanent) or 302 (temporary) redirects.

This feature manages dynamic "Single Redirects". (Single Redirects can be
static or dynamic but DNSControl only maintains dynamic redirects).

Cloudflare documentation: <https://developers.cloudflare.com/rules/url-forwarding/single-redirects/>

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  CF_SINGLE_REDIRECT("name", 301, "when", "then"),
  CF_SINGLE_REDIRECT('redirect www.example.com', 301, 'http.host eq "www.example.com"', 'concat("https://otherplace.com", http.request.uri.path)'),
  CF_SINGLE_REDIRECT('redirect yyy.example.com', 301, 'http.host eq "yyy.example.com"', 'concat("https://survey.stackoverflow.co", "")'),
END);
```
{% endcode %}

The fields are:

* name: The name (basically a comment, but it must be unique)
* code: Either 301 (permanent) or 302 (temporary) redirects. May be a number or string.
* when: What Cloudflare sometimes calls the "rule expression".
* then: The replacement expression.

{% hint style="info" %}
**NOTE**: The features [`CF_REDIRECT`](CF_REDIRECT.md) and [`CF_TEMP_REDIRECT`](CF_TEMP_REDIRECT.md) generate `CF_SINGLE_REDIRECT` if enabled in [`CLOUDFLAREAPI`](../../provider/cloudflareapi.md).
{% endhint %}
