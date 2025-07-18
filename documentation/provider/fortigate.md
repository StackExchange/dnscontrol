# FortiGate DNS Provider

This DNS provider lets you manage DNS zones hosted on a Fortinet FortiGate device via its REST API.

## Configuration

The provider is configured using the following environment variables or entries in `creds.json`:

- `FORTIGATE_HOST`: The FortiGate host or IP address (e.g. `https://192.168.1.1`)
- `FORTIGATE_TOKEN`: API token with appropriate DNS permissions
- `FORTIGATE_VDOM`: (optional) Specify the virtual domain (default: `root`)
- `FORTIGATE_INSECURE_TLS`: (optional) Set to `true` to disable SSL certificate verification (useful for self-signed certs)
- `FORTIGATE_DEBUG_HTTP`: (optional) Set to `true` to log raw HTTP requests/responses

Example `creds.json` entry:

```json
{
  "FORTIGATE": {
    "host": "https://192.168.1.1",
    "token": "your-api-token",
    "vdom": "root",
    "insecure_tls": true,
    "debug_http": true
  }
}
```

## Usage

To use this provider in a `dnsconfig.js`:

```javascript
D("example.com", REG_NONE, DnsProvider("FORTIGATE"),
  A("www", "192.0.2.1"),
  CNAME("blog", "external.example.net."),
  MX("@", 10, "mail.example.com."),
  NS("@", "ns1.example.net.")
)
```

### Record Status (Enable/Disable)

FortiGate supports disabling DNS records (setting them as `status: disable`).  
This provider maps that setting to record metadata in dnscontrol.

To disable a record, set the following metadata key:

```javascript
A("disabledhost", "192.0.2.123", { metadata: { fortigate_status: "disable" } })
```

✅ Supported record types: `A`, `AAAA`, `CNAME`, `NS`, `MX`

## Caveats

- ✅ **NS and MX records are supported, with limitations:**  
  - Only apex records (hostname `"@"`) are supported.  
  - MX records must have a valid hostname (not `"."`).  
  - FortiGate does not enforce priority uniqueness or ordering.

- ❌ **PTR records are not supported.**  
  FortiGate stores reverse DNS data unconventionally. PTR records are excluded to prevent inconsistencies.

- ❌ **TXT records are not supported.**  
  The FortiGate API does not currently allow TXT records.

- ❌ **Wildcard records (`*`) are not supported.**  
  The FortiGate DNS engine does not support wildcard entries.


## Development notes

This provider uses the FortiGate REST API (`/api/v2/cmdb/system/dns-database`) to manage zones and DNS entries. It operates on the **"shadow" DNS database**, assuming zones are configured in **primary mode** (not forwarded). It automatically creates zones if they do not exist.

Debug logging of HTTP traffic can be enabled with the `debug_http` flag.