---
name: FORSALE_BUILDER
parameters:
  - label
  - opts
  - modifiers...
parameter_types:
  label: string
  opts: "{ fcod?: string | string[]; ftxt?: string | string[]; furi?: string | string[]; fval?: string | string[] }?"
  "modifiers...": RecordModifier[]
---

`FORSALE_BUILDER` creates the `_for-sale` [`TXT`](TXT.md) records of
[RFC 10023](https://www.rfc-editor.org/rfc/rfc10023), which signal that a
registered domain name is available for purchase.

The first parameter is the label of the name that is for sale, usually `"@"`.
The records are created at `_for-sale` below that label.

## Example

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  FORSALE_BUILDER("@", {
      furi: "https://example.com/buy",
      fval: "EUR999",
  }, TTL("1h")),
);
```
{% endcode %}

This yields the following records:

```text
_for-sale  IN  TXT  "v=FORSALE1;furi=https://example.com/buy"
_for-sale  IN  TXT  "v=FORSALE1;fval=EUR999"
```

Without options, `FORSALE_BUILDER("@")` creates the single record
`"v=FORSALE1;"`, which says that the domain name is for sale without saying
anything else. RFC 10023 recommends giving at least one content tag.

## Parameters

Each option is a string or a list of strings. The RFC allows only one content
tag per record, so the builder creates one record per value, in the order
`fcod`, `ftxt`, `furi`, `fval`, and within an option in the order given.

| Option | Description |
|---|---|
| `fcod` | A code that only parties who agreed on its meaning understand, for example a registry and its registrars: `"EXCO-S2lscm95IHdhcyBoZXJl"`. |
| `ftxt` | A short text for people: `"Call for info."`. |
| `furi` | One absolute URI (or IRI) with more information, preferably `https:`, `mailto:` or `tel:`: `"https://example.com/buy"`. Spaces must be percent-encoded (`%20`). |
| `fval` | An asking price: an upper case currency code followed by the amount, with an optional decimal point and no separators: `"USD750"`, `"BTC0.000010"`. |

Every value must be at least one and at most 239 octets long, so that the record
fits in a single TXT string of 255 octets as the RFC requires. The same value
may not be given twice for one option.

The builder refuses names below `.arpa` and wildcards, where the RFC says the
records have no meaning.

RFC 10023 recommends a TTL of one hour or less, so that a domain name that is no
longer for sale, or a changed price, does not stay in caches for long.

When the domain name is no longer for sale, remove the `FORSALE_BUILDER` call.
