This is a read-only/monitoring "registrar". It does a DNS NS lookup to confirm the nameserver servers are correct. This "registrar" is unable to update/correct the NS servers but will alert you if they are incorrect. A common use of this provider is when the domain is with a registrar that does not have an API.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DNSOVERHTTPS`.

```json
{
  "dohdefault": {
    "TYPE": "DNSOVERHTTPS"
  }
}
```

The DNS-over-HTTPS provider defaults to using Google Public DNS however you may configure an alternative RFC 8484 DoH provider using the `host` parameter.

Example:

```json
{
  "dohcloudflare": {
    "TYPE": "DNSOVERHTTPS",
    "host": "cloudflare-dns.com"
  }
}
```

Some common DoH providers are:

* `cloudflare-dns.com` ([Cloudflare](https://developers.cloudflare.com/1.1.1.1/dns-over-https))
* `9.9.9.9` ([Quad9](https://www.quad9.net/about/))
* `dns.google` ([Google Public DNS](https://developers.google.com/speed/public-dns/docs/doh))

## Metadata
This provider does not recognize any special metadata fields unique to DOH.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_MONITOR = NewRegistrar("dohcloudflare");

D("example.com", REG_MONITOR,
  NAMESERVER("ns1.example.com."),
  NAMESERVER("ns2.example.com."),
);
```

{% hint style="info" %}
**NOTE**: This checks the NS records via a DNS query.  It does not check the
registrar's delegation (i.e. the `Name Server:` field in whois). In theory
these are the same thing but there may be situations where they are not.
{% endhint %}
