# MikroTik RouterOS DNS Provider

This DNS provider manages DNS static entries on a MikroTik RouterOS device via its REST API.

## Supported Features

- `dnscontrol get-zones` is supported. Zones are auto-detected by grouping record names by their effective domain suffix.
- Supported record types: `A`, `AAAA`, `CNAME`, `MX`, `NS`, `SRV`, `TXT`
- Custom record types:
  - `MIKROTIK_FWD` — RouterOS FWD entries for conditional DNS forwarding with address list population.
  - `MIKROTIK_NXDOMAIN` — RouterOS NXDOMAIN entries (respond with NXDOMAIN for matching queries).
  - `MIKROTIK_FORWARDER` — RouterOS DNS forwarder entries (managed via the synthetic `_forwarders.mikrotik` zone).

## Configuration

The provider is configured using entries in `creds.json`:

- `host`: RouterOS REST API endpoint (e.g. `http://192.168.88.1:8080`)
- `username`: RouterOS user with API access
- `password`: Password for the user
- `zonehints`: (optional) Comma-separated list of zone names to help identify zones with 3+ labels (e.g. `internal.corp.local,home.arpa`)

Example `creds.json` entry:

```json
{
  "mikrotik": {
    "TYPE": "MIKROTIK",
    "host": "http://192.168.88.1:8080",
    "username": "admin",
    "password": "secret",
    "zonehints": "internal.corp.local,home.arpa"
  }
}
```

### Zone Detection

RouterOS has no native zone concept — DNS static entries are a flat list. The provider groups records into zones by their domain suffix:

1. If `zonehints` is configured, records are matched against hints (longest match wins).
2. Otherwise, `publicsuffix.EffectiveTLDPlusOne` is used for public TLDs.
3. For private TLDs (e.g. `.local`), the last two labels are used as a fallback.

Use `zonehints` when you have zones with 3+ labels (e.g. `h.example.com` as a separate zone from `example.com`).

## Record Metadata

All record types support the following metadata keys:

| Key               | Type   | Description                                                        |
|-------------------|--------|--------------------------------------------------------------------|
| `match_subdomain` | string | Set to `"true"` to enable RouterOS subdomain matching.             |
| `regexp`          | string | RouterOS regexp for matching queries.                              |
| `address_list`    | string | RouterOS address list to populate with resolved addresses.         |
| `comment`         | string | Comment stored on the RouterOS record.                             |

### MIKROTIK_FWD

Forward DNS queries to a specified server. Commonly used for conditional forwarding with address list population.

```javascript
MIKROTIK_FWD("@", "8.8.8.8", {match_subdomain: "true", address_list: "vpn-list"})
```

### MIKROTIK_NXDOMAIN

Return NXDOMAIN for matching queries.

```javascript
MIKROTIK_NXDOMAIN("ads.example.com", "")
```

### MIKROTIK_FORWARDER

Manage RouterOS DNS forwarder entries via the synthetic `_forwarders.mikrotik` zone.
Additional metadata keys for forwarders:

| Key                | Type   | Description                                        |
|--------------------|--------|----------------------------------------------------|
| `doh_servers`      | string | DoH server URLs for this forwarder.                |
| `verify_doh_cert`  | string | Set to `"true"` to verify DoH certificate.         |

```javascript
D("_forwarders.mikrotik", REG_CHANGEME,
  DnsProvider(DSP_MIKROTIK),
  MIKROTIK_FORWARDER("example.com", "1.1.1.1")
)
```

## Usage

```javascript
var DSP_MIKROTIK = NewDnsProvider("mikrotik", "MIKROTIK");
var REG_CHANGEME = NewRegistrar("none");

D("example.com", REG_CHANGEME,
  {no_ns: "true"},
  DnsProvider(DSP_MIKROTIK),
  A("www", "192.0.2.1"),
  AAAA("www", "2001:db8::1"),
  CNAME("blog", "www.example.com."),
  MX("@", 10, "mail.example.com."),
  MIKROTIK_FWD("@", "8.8.8.8", {match_subdomain: "true", address_list: "vpn-list"})
)
```

**Note:** RouterOS does not expose nameservers, so `{no_ns: "true"}` should be set on all domains to suppress the "Skipping registrar" warning.

## Activation

The RouterOS REST API must be enabled on the device.

### Enable REST API (RouterOS 7.x)

Via the RouterOS CLI (SSH or terminal):

```
/ip/service/set www-ssl disabled=no
/certificate/add name=local-cert common-name=router
/ip/service/set www-ssl certificate=local-cert
```

Or for HTTP (not recommended for production):

```
/ip/service/set www disabled=no port=8080
```

### Create a dedicated user

```
/user/add name=dnscontrol password=secret group=full
```

For read-only preview, use `group=read`.

## Caveats

- **No native zone concept.** Zones are inferred from record names. Use `zonehints` for multi-label private zones.
- **MX records with target `.` (null MX) are rejected** by the audit system.
- **Dynamic and disabled records are ignored** during zone enumeration and record fetching.
- **TTL values** are stored in RouterOS duration format (e.g. `1d`, `1h30m`) and converted automatically.

## Development Notes

This provider uses the RouterOS REST API endpoints:
- `/rest/ip/dns/static` — for DNS static records (A, AAAA, CNAME, FWD, MX, NS, NXDOMAIN, SRV, TXT)
- `/rest/ip/dns/forwarders` — for DNS forwarder entries

Records are compared using `diff2.ByRecord()` with custom comparison functions that include metadata fields (`match_subdomain`, `regexp`, `address_list`, `comment`) so that metadata-only changes are properly detected.
