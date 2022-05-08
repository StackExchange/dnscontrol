---
name: Hetzner DNS Console
title: Hetzner DNS Console
layout: default
jsId: HETZNER
---

# Hetzner DNS Console Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `HETZNER`
along with a [Hetzner API Key](https://dns.hetzner.com/settings/api-token).

Example:

```json
{
  "hetzner": {
    "TYPE": "HETZNER",
    "api_key": "your-api-key"
  }
}
```

## Metadata

This provider does not recognize any special metadata fields unique to Hetzner
 DNS Console.

## Usage

An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("none");
var DSP_HETZNER = NewDnsProvider("hetzner");

D("example.tld", REG_NONE, DnsProvider(DSP_HETZNER),
    A("test", "1.2.3.4")
);
```

## Activation

Create a new API Key in the
[Hetzner DNS Console](https://dns.hetzner.com/settings/api-token).

## Caveats

### SOA

Hetzner DNS Console does not allow changing the SOA record via their API.
There is an alternative method using an import of a full BIND file, but this
 approach does not play nice with incremental changes or ignored records.
At this time you cannot update SOA records via DNSControl.

### Rate Limiting

Hetzner is rate limiting requests in multiple tiers: per Hour, per Minute and
 per Second.

Depending on how many requests you are planning to perform, you can adjust the
 delay between requests in order to stay within your quota.

The setting `optimize_for_rate_limit_quota` controls this behavior and accepts
 a case-insensitive value of
- `Hour`
- `Minute`
- `Second`

The default for `optimize_for_rate_limit_quota` is `Second`.

Example: Your per minute quota is 60 requests and in your settings you
 specified `Minute`. DNSControl will perform at most one request per second.
 DNSControl will emit a warning in case it breaches the next quota.

In your `creds.json` for all `HETZNER` provider entries:

```json
{
  "hetzner": {
    "TYPE": "HETZNER",
    "api_key": "your-api-key",
    "optimize_for_rate_limit_quota": "Minute"
  }
}
```

Every response from the Hetzner DNS Console API includes your limits:

```bash
curl --silent --include \
    --header 'Auth-API-Token: ...' \
    https://dns.hetzner.com/api/v1/zones \
  | grep x-ratelimit-limit
x-ratelimit-limit-second: 3
x-ratelimit-limit-minute: 42
x-ratelimit-limit-hour: 1337
```

Every DNSControl invocation starts from scratch in regard to rate-limiting.
In case you are frequently invoking DNSControl, you will likely hit a limit for
 any first request.
You can either use an out-of-bound delay (e.g. `$ sleep 1`), or specify
 `start_with_default_rate_limit` in the settings of the provider.
With `start_with_default_rate_limit` DNSControl uses a quota equivalent to
 `x-ratelimit-limit-second: 1` until it could parse the actual quota from an
 API response.

In your `creds.json` for all `HETZNER` provider entries:

```json
{
  "hetzner": {
    "TYPE": "HETZNER",
    "api_key": "your-api-key",
    "start_with_default_rate_limit": "true"
  }
}
```
