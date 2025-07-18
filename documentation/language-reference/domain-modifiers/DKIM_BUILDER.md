---
name: DKIM_BUILDER
parameters:
  - label
  - selector
  - pubkey
  - flags
  - hashtypes
  - keytype
  - servicetypes
  - note
  - ttl
parameters_object: true
parameter_types:
  label: string?
  selector: string
  pubkey: string
  flags: string[]?
  hashtypes: string[]?
  keytype: string?
  servicetypes: string[]?
  note: string?
  ttl: Duration?
---

DNSControl contains a `DKIM_BUILDER` which can be used to simply create
DKIM policies for your domains.


## Example

### Simple example

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DKIM_BUILDER({
    selector: "s1",
    pubkey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDC5/z4L"
  }),
);
```
{% endcode %}

This yield the following record:

```text
s1._domainkey   IN  TXT "v=DKIM1; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDC5/z4L"
```

### Advanced example

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  DKIM_BUILDER({
    label: "alerts",
    selector: "k2",
    pubkey: "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDC5/z4L",
    flags: ['y'],
    hashtypes: ['sha256'],
    keytype: 'rsa',
    servicetypes: ['email'],
    ttl: 150
  }),
);
```
{% endcode %}

This yields the following record:

```text

k2._domainkey.alerts    IN  TXT "v=DKIM1; k=rsa; s=email; t=y; h=sha256; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDC5/z4L" ttl=150

```

### Parameters

* `label:` The DNS label for the DKIM record (`[selector]._domainkey` prefix is added; default: `'@'`)
* `selector:` Selector used for the label. e.g. `s1` or `mail`
* `pubkey:` Public key `p` to be used for DKIM.
* `keytype:` Key type `k`. Defaults to `'rsa'` if omitted (optional)
* `flags:` Which types `t` of flags to activate, ie. 'y' and/or 's'. Array, defaults to 's' (optional)
* `hashtypes:` Acceptable hash algorithms `h` (optional)
* `servicetypes:` Record-applicable service types (optional)
* `note:` Note field `n` for admins. Avoid if possible to keep record length short. (optional)
* `ttl:` Input for `TTL` method (optional)

### Caveats

* DKIM (TXT) records are automatically split using `AUTOSPLIT`.
