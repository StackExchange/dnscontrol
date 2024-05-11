## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `HETZNER`
along with a [Hetzner API Key](https://dns.hetzner.com/settings/api-token).

Example:

{% code title="creds.json" %}
```json
{
  "hetzner": {
    "TYPE": "HETZNER",
    "api_key": "your-api-key"
  }
}
```
{% endcode %}

## Metadata

This provider does not recognize any special metadata fields unique to Hetzner
 DNS Console.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_HETZNER = NewDnsProvider("hetzner");

D("example.com", REG_NONE, DnsProvider(DSP_HETZNER),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation

Create a new API Key in the
[Hetzner DNS Console](https://dns.hetzner.com/settings/api-token).

## Caveats

### CAA

As of June 2022, the Hetzner DNS Console API does not accept spaces in CAA
 records.
```text
0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

Removing the spaces might still work for any consumer of the record.
```text
0 issue "letsencrypt.org;validationmethods=dns-01;accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

### SOA

Hetzner DNS Console does not allow changing the SOA record via their API.
There is an alternative method using an import of a full BIND file, but this
 approach does not play nice with incremental changes or ignored records.
At this time you cannot update SOA records via DNSControl.

### Rate Limiting

Hetzner is rate limiting requests quite heavily.

The rate limit and remaining quota is advertised in the API response headers.

DNSControl will burst through half of the quota, and then it spreads the
 requests evenly throughout the remaining window. This allows you to move fast
 and be able to revert accidental changes to the DNS config in a timely manner.

Every response from the Hetzner DNS Console API includes your limits:

```shell
curl --silent --include \
    --header 'Auth-API-Token: ...' \
    https://dns.hetzner.com/api/v1/zones

Access-Control-Allow-Origin *
Content-Type application/json; charset=utf-8
Date Sat, 01 Apr 2023 00:00:00 GMT
Ratelimit-Limit 42
Ratelimit-Remaining 33
Ratelimit-Reset 7
Vary Origin
X-Ratelimit-Limit-Minute 42
X-Ratelimit-Remaining-Minute 33
```
With the above values, DNSControl will not delay the next 12 requests (until it
 hits `Ratelimit-Remaining: 21 # 42/2`) and then slow down requests with a
 delay of `7s/22 ≈ 300ms` between requests (about 3 requests per second).
Performing these 12 requests might take longer than 7s, at which point the
 quota resets and DNSControl will burst through the quota again.

DNSControl will retry rate-limited requests (status 429) and respect the
 advertised `Retry-After` delay.
