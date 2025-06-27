
# FortiGate DNS Provider

This DNS provider lets you manage DNS zones hosted on a Fortinet FortiGate device via its REST API.

## Configuration

The provider is configured using the following environment variables:

- `FORTIGATE_HOST`: The FortiGate host or IP address (e.g. `https://192.168.1.1`)
- `FORTIGATE_TOKEN`: API token with appropriate DNS permissions
- `FORTIGATE_VDOM`: (optional) Specify the virtual domain (default: `root`)
- `FORTIGATE_INSECURE_TLS`: (optional) Set to `true` to disable SSL certificate verification (useful for self-signed certs)

Example `creds.json` entry:

{% code title="creds.json" %}
```json
{
  "FORTIGATE": {
    "host": "https://192.168.1.1",
    "token": "your-api-token",
    "vdom": "root",
    "insecure_tls": true
  }
}
```
{% endcode %}

## Usage

To use this provider in a `dnsconfig.js`:

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_NONE, DnsProvider("FORTIGATE"),
  A("www", "192.0.2.1"),
  CNAME("blog", "external.example.net.")
)
```
{% endcode %}

⚠️ TXT records are not supported. See caveats below.

## Caveats

- ❌ **PTR records are not supported.**

  FortiGate does not follow the standard DNS convention of managing `in-addr.arpa` or `ip6.arpa` zones for reverse DNS. Instead, PTR entries are stored in regular forward zones, and this behavior is incompatible with how `dnscontrol` models reverse zones. Because of this mismatch, PTR support is intentionally omitted to avoid unexpected behavior or broken state synchronization.

- ❌ **NS and MX records are not supported.**

  FortiGate does not support fully functional `NS` or `MX` record types in its DNS configuration system.

- ❌ **TXT records are not supported.**

  The FortiGate DNS interface does not currently expose support for TXT records via the API.

- ❌ **Wildcard records (`*`) are not supported.**

  FortiGate DNS does not support wildcard records.

- ✅ Supported record types: `A`, `AAAA`, `CNAME`.

## Development Notes

This provider uses the FortiGate REST API (`/api/v2/cmdb/system/dns-database`) to manage zones and DNS entries. It assumes you are managing the **"shadow" view** and expects zones to be configured in **primary mode**.