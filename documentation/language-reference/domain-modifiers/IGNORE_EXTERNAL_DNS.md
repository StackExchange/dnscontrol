---
name: IGNORE_EXTERNAL_DNS
---

`IGNORE_EXTERNAL_DNS` makes DNSControl automatically detect and ignore DNS records
managed by Kubernetes external-dns.

## Background

[External-dns](https://github.com/kubernetes-sigs/external-dns) is a popular
Kubernetes controller that synchronizes exposed Kubernetes Services and Ingresses
with DNS providers. It creates DNS records automatically based on annotations on
your Kubernetes resources.

External-dns uses TXT records to track ownership of the DNS records it manages.
These TXT records contain metadata in this format:

```
"heritage=external-dns,external-dns/owner=<owner-id>,external-dns/resource=<resource>"
```

When you have both DNSControl and external-dns managing the same DNS zone, conflicts
can occur. DNSControl will try to delete records created by external-dns, and
external-dns will recreate them, leading to an endless update cycle.

## How it works

When `IGNORE_EXTERNAL_DNS` is enabled, DNSControl will:

1. Scan existing TXT records for the external-dns heritage marker (`heritage=external-dns`)
2. Parse the TXT record name to determine which DNS record it manages
3. Automatically ignore both the TXT ownership record and the corresponding DNS record

External-dns creates TXT records with prefixes based on record type:
- `a-<name>` for A records
- `aaaa-<name>` for AAAA records  
- `cname-<name>` for CNAME records
- `ns-<name>` for NS records
- `mx-<name>` for MX records
- `srv-<name>` for SRV records

For example, if external-dns creates an A record at `myapp.example.com`, it will
also create a TXT record at `a-myapp.example.com` containing the heritage information.

## Usage

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  IGNORE_EXTERNAL_DNS,
  // Your static DNS records managed by DNSControl
  A("www", "1.2.3.4"),
  A("mail", "1.2.3.5"),
  MX("@", 10, "mail"),
  // Records created by external-dns (from Kubernetes Ingresses/Services)
  // will be automatically detected and ignored
);
```
{% endcode %}

## Example scenario

Suppose you have:
- A Kubernetes cluster running external-dns with `--txt-owner-id=my-cluster`
- An Ingress resource that creates an A record for `myapp.example.com` pointing to `10.0.0.1`

External-dns will create:
1. An A record: `myapp.example.com` → `10.0.0.1`
2. A TXT record: `a-myapp.example.com` → `"heritage=external-dns,external-dns/owner=my-cluster,external-dns/resource=ingress/default/myapp"`

With `IGNORE_EXTERNAL_DNS` enabled, DNSControl will:
- Detect the TXT record at `a-myapp.example.com` as an external-dns ownership record
- Ignore both the TXT record and the A record at `myapp.example.com`
- Only manage the records you explicitly define in your `dnsconfig.js`

## Comparison with other options

| Feature | Use case |
|---------|----------|
| `IGNORE_EXTERNAL_DNS` | Automatically ignore all external-dns managed records |
| `IGNORE("*.k8s", "A,AAAA,CNAME,TXT")` | Ignore records under a specific subdomain pattern |
| `NO_PURGE` | Don't delete any records (less precise, records may accumulate) |

## Caveats

### TXT Registry Format

This feature relies on external-dns's [TXT registry](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/registry/txt.md),
which is the default registry type. The TXT record content format is well-documented:

```
"heritage=external-dns,external-dns/owner=<owner-id>,external-dns/resource=<resource>"
```

This feature detects the `heritage=external-dns` marker in TXT records to identify
external-dns managed records.

### Default Prefix Required

This feature works with external-dns's **default** TXT record naming convention,
which uses the `--txt-prefix="%{record_type}-"` format (the default since v0.16+).
This creates TXT records like:

- `a-myapp.example.com` for A records
- `cname-api.example.com` for CNAME records

If you've configured external-dns with a **custom** `--txt-prefix` or `--txt-suffix`,
those records may not be detected correctly. For example:

- `--txt-prefix="extdns-"` would create `extdns-myapp.example.com` (not detected)
- `--txt-suffix="-.extdns"` would create `myapp-.extdns.example.com` (not detected)

If you need support for custom prefixes/suffixes, please open an issue.

### Unsupported Registries

External-dns supports multiple registry types. This feature **only** supports:

- ✅ **TXT registry** (default) - Stores metadata in TXT records

The following registries are **not supported**:

- ❌ **DynamoDB registry** - Stores metadata in AWS DynamoDB
- ❌ **AWS-SD registry** - Stores metadata in AWS Service Discovery  
- ❌ **noop registry** - No metadata persistence

### Legacy TXT Format

External-dns versions prior to v0.16 created TXT records without the record type
prefix (e.g., `myapp.example.com` instead of `a-myapp.example.com`). This legacy
format is partially supported but may match more records than intended.

## See also

* [`IGNORE`](IGNORE.md) for manually ignoring specific records with glob patterns
* [`NO_PURGE`](NO_PURGE.md) for preventing deletion of all unmanaged records
* [External-dns documentation](https://github.com/kubernetes-sigs/external-dns)
