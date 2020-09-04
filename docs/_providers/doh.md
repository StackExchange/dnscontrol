---
name: DNS-over-HTTPS
title: DNS-over-HTTPS Provider
layout: default
jsId: DNSOVERHTTPS
---
# DNS-over-HTTPS Provider

This is a read-only/monitoring "registrar". It does a DNS NS lookup to confirm the nameserver servers are correct. This "registrar" is unable to update the NS servers but will alert you if they are incorrect. A common use of this provider is when the domain is with a registrar that does not have an API.

## Configuration
The DNS-over-HTTPS provider does not require anything in `creds.json`. By default, it uses Google Public DNS however you may configure an alternative RFC 8484 DoH provider.

{% highlight json %}
{
  "DNS-over-HTTPS": {
    "host": "cloudflare-dns.com"
  }
}
{% endhighlight %}

Some common DoH providers are `cloudflare-dns.com` [Cloudflare](https://developers.cloudflare.com/1.1.1.1/dns-over-https), `9.9.9.9` [Quad9](https://www.quad9.net/about/), and `dns.google` [Google Public DNS](https://developers.google.com/speed/public-dns/docs/doh)

## Metadata
This provider does not recognize any special metadata fields unique to Internet.bs.

## Usage
Example Javascript:

{% highlight js %}
var REG_MONITOR = NewRegistrar('DNS-over-HTTPS', 'DNSOVERHTTPS');

D("example.com", REG_MONITOR,
  NNAMESERVER("ns1.example.com."),
  NNAMESERVER("ns2.example.com."),
);
{% endhighlight %}
