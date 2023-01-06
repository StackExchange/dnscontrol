---
title: Service Providers
---
# Service Providers

## Provider Features {#features}

The table below shows various features supported, or not supported by DNSControl providers.
Underlined items have tooltips for more detailed explanation. This table is automatically generated
from metadata supplied by the provider when they register themselves inside dnscontrol.

An empty space may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.
<br/>
<br/>

{% include matrix.html %}


### Providers with "official support"

Official support means:

* New releases will block if any of these providers do not pass integration tests.
* The DNSControl maintainers prioritize fixing bugs in these providers (though we gladly accept PRs).
* New features will work on these providers (unless the provider does not support it).
* StackOverflow maintains test accounts with those providers for running integration tests.

Providers in this category and their maintainers are:

|Name|Maintainer|
|---|---|
|`AZURE_DNS`|@vatsalyagoel|
|`BIND`|@tlimoncelli|
|`GCLOUD`|@riyadhalnur|
|`NAMEDOTCOM`|@tlimoncelli|

### Providers with "contributor support"

The other providers are supported by community members, usually the
original contributor.

Due to the large number of DNS providers in the world, the DNSControl
team can not support and test all providers.  Test frameworks are
provided to help community members support their code independently.

Expectations of maintainers:

* Maintainers are expected to support their provider and/or find a new maintainer.
* Maintainers should set up test accounts and periodically verify that all tests pass (`pkg/js/parse_tests` and `integrationTest`).
* Contributors are encouraged to add new tests and refine old ones. (Test-driven development is encouraged.)
* Bugs will be referred to the maintainer or their designate.
* Maintainers must be responsible to bug reports and PRs.  If a maintainer is unresponsive for more than 2 months, we will consider disabling the provider.  First we will put out a call for new maintainer. If noboby volunteers, the provider will be disabled.

Providers in this category and their maintainers are:

|Name|Maintainer|
|---|---|
|`AXFRDDNS`|@hnrgrgr|
|`AKAMAIEDGEDNS`|@svernick|
|`CLOUDNS`|@pragmaton|
|`CLOUDFLAREAPI`|@tresni|
|`CSCGLOBAL`|@Air-New-Zealand|
|`DESEC`|@D3luxee|
|`DIGITALOCEAN`|@Deraen|
|`DNSOVERHTTPS`|@mikenz|
|`DNSIMPLE`|@onlyhavecans|
|`DNSMADEEASY`|@vojtad|
|`DOMAINNAMESHOP`|@SimenBai|
|`EASYNAME`|@tresni|
|`EXOSCALE`|@pierre-emmanuelJ|
|`GANDI_V5`|@TomOnTime|
|`GCORE`|@xddxdd|
|`HEDNS`|@rblenkinsopp|
|`HETZNER`|@das7pad|
|`HEXONET`|@papakai|
|`HOSTINGDE`|@membero|
|`INTERNETBS`|@pragmaton|
|`INWX`|@svenpeter42|
|`LINODE`|@koesie10|
|`NAMECHEAP`|VOLUNTEER NEEDED|
|`NETCUP`|@kordianbruck|
|`NETLIFY`|@SphericalKat|
|`NS1`|@costasd|
|`OPENSRS`|@pierre-emmanuelJ|
|`ORACLE`|@kallsyms|
|`OVH`|@masterzen|
|`PACKETFRAME`|@hamptonmoore|
|`POWERDNS`|@jpbede|
|`RWTH`|@MisterErwin|
|`ROUTE53`|@tresni|
|`SOFTLAYER`|@jamielennox|
|`TRANSIP`|@blackshadev|
|`VULTR`|@pgaskin|

### Requested providers

We have received requests for the following providers. If you would like to contribute
code to support this provider, please re-open the issue. We'd be glad to help in any way.

* [Support Gandi as a registrar](https://github.com/StackExchange/dnscontrol/issues/87) (#87)
* [Provider request: GoDaddy](https://github.com/StackExchange/dnscontrol/issues/145) (#145)
* [Provider request: NameSilo](https://github.com/StackExchange/dnscontrol/issues/220) (#220)
* [OpenSRS: Add Registrar and DSP support](https://github.com/StackExchange/dnscontrol/issues/272) (#272)
* [Provider request: Oracle Cloud Infrastructure](https://github.com/StackExchange/dnscontrol/issues/419) (#419)
* [Provider request: Alibaba cloud ](https://github.com/StackExchange/dnscontrol/issues/420)(#420)
* [Provider request: Netlify](https://github.com/StackExchange/dnscontrol/issues/714) (#714)
* [Provider request: Hetzner](https://github.com/StackExchange/dnscontrol/issues/715) (#715)
* [Provider request: CSC Global](https://github.com/StackExchange/dnscontrol/issues/815) (#815)
* [Provider request: Porkbun](https://github.com/StackExchange/dnscontrol/issues/1295) (#1295)

### In progress providers

These requests have an *open* issue, which indicates somebody is actively working on it. Feel free to follow the issue, or pitch in if you think you can help.

* [Provider Request: knot-dns](https://github.com/StackExchange/dnscontrol/issues/436) (#436)
* [Provider request: Constellix (DNSMadeEasy)](https://github.com/StackExchange/dnscontrol/issues/842) (#842)
* [Provider request: Joker.com](https://github.com/StackExchange/dnscontrol/issues/854) (#854)
* [Provider request: RcodeZero](https://github.com/StackExchange/dnscontrol/issues/884) (#884)
* [New Feature Request: Customer-Accessible API for EnCirca](https://github.com/StackExchange/dnscontrol/issues/1048) (#1048)
* [Provider request: Infoblox/NIOS](https://github.com/StackExchange/dnscontrol/issues/1077) (#1077)
* [EU.ORG Registrar Support](https://github.com/StackExchange/dnscontrol/issues/1176) (#1176)
* [DNS Provider: 1984Hosting Support](https://github.com/StackExchange/dnscontrol/issues/1251) (#1251)
* [PROVIDER REQUEST: coredns](https://github.com/StackExchange/dnscontrol/issues/1284) (#1284)
* [Provider request: AutoDNS](https://github.com/StackExchange/dnscontrol/issues/1323) (#1323)

### Providers with open PRs

These providers have an open PR with (potentially) working code. They may be ready to merge, or may have blockers. See issue and PR for details.
