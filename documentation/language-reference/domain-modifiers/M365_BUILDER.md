---
name: M365_BUILDER
parameters:
  - label
  - mx
  - autodiscover
  - dkim
  - skypeForBusiness
  - mdm
  - domainGUID
  - initialDomain
parameters_object: true
parameter_types:
  label: string?
  mx: boolean?
  autodiscover: boolean?
  dkim: boolean?
  skypeForBusiness: boolean?
  mdm: boolean?
  domainGUID: string?
  initialDomain: string?
---

DNSControl offers a `M365_BUILDER` which can be used to simply set up Microsoft 365 for a domain in an opinionated way.

It defaults to a setup without support for legacy Skype for Business applications.
It doesn't set up SPF or DMARC. See [`SPF_BUILDER`](SPF_BUILDER.md) and [`DMARC_BUILDER`](DMARC_BUILDER.md).

## Example

### Simple example

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  M365_BUILDER({
      initialDomain: "example.onmicrosoft.com",
  }),
END);
```
{% endcode %}

This sets up `MX` records, Autodiscover, and DKIM.

### Advanced example

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_MY_PROVIDER, DnsProvider(DSP_MY_PROVIDER),
  M365_BUILDER({
      label: "test",
      mx: false,
      autodiscover: false,
      dkim: false,
      mdm: true,
      domainGUID: "test-example-com", // Can be automatically derived in this case, if example.com is the context.
      initialDomain: "example.onmicrosoft.com",
  }),
END);
```
{% endcode %}

This sets up Mobile Device Management only.

### Parameters

* `label` The label of the Microsoft 365 domain, useful if it is a subdomain (default: `"@"`)
* `mx` Set an `MX` record? (default: `true`)
* `autodiscover` Set Autodiscover `CNAME` record? (default: `true`)
* `dkim` Set DKIM `CNAME` records? (default: `true`)
* `skypeForBusiness` Set Skype for Business/Microsoft Teams records? (default: `false`)
* `mdm` Set Mobile Device Management records? (default: `false`)
* `domainGUID` The GUID of _this_ Microsoft 365 domain (default: `<label>.<context>` with `.` replaced by `-`, no default if domain contains dashes)
* `initialDomain` The initial domain of your Microsoft 365 tenant/account, ends in `onmicrosoft.com`
