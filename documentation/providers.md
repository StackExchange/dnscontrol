# Service Providers

## Provider Features

The table below shows various features supported, or not supported by DNSControl providers.
This table is automatically generated from metadata supplied by the provider when they register themselves inside dnscontrol.

An empty space may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.

| Provider name | Official Support | DNS Provider | Registrar | ALIAS | AUTODNSSEC | CAA | PTR | NAPTR | SOA | SRV | SSHFP | TLSA | DS | dual host | create-domains | NO_PURGE | get-zones |
| ------------- | ---------------- | ------------ | --------- | ----- | ---------- | --- | --- | ----- | --- | --- | ----- | ---- | -- | --------- | -------------- | -------- | --------- |
| `AKAMAIEDGEDNS` | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ |
| `AUTODNS` | ❌ | ✅ | ❌ | ✅ | ❔ | ❌ | ❌ | ❔ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| `AXFRDDNS` | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❌ | ❌ | ❌ | ❌ |
| `AZURE_DNS` | ✅ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ❌ | ❔ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ |
| `BIND` | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| `CLOUDFLAREAPI` | ✅ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❔ | ❌ | ✅ | ✅ | ✅ |
| `CLOUDNS` | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| `CSCGLOBAL` | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ✅ |
| `DESEC` | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ |
| `DIGITALOCEAN` | ❌ | ✅ | ❌ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| `DNSIMPLE` | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| `DNSMADEEASY` | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `DNSOVERHTTPS` | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| `DOMAINNAMESHOP` | ❌ | ✅ | ❌ | ❔ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ |
| `EASYNAME` | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| `EXOSCALE` | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ❌ | ❔ | ❌ | ❌ | ✅ | ❔ |
| `GANDI_V5` | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ✅ |
| `GCLOUD` | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ |
| `GCORE` | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `HEDNS` | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `HETZNER` | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❔ | ❔ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `HEXONET` | ❌ | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ |
| `HOSTINGDE` | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `INTERNETBS` | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| `INWX` | ❌ | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ |
| `LINODE` | ❌ | ✅ | ❌ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ✅ |
| `MSDNS` | ✅ | ✅ | ❌ | ❌ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ✅ |
| `NAMECHEAP` | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❌ | ❌ | ✅ |
| `NAMEDOTCOM` | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ | ✅ |
| `NETCUP` | ❌ | ✅ | ❌ | ❔ | ❔ | ✅ | ❌ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ❌ |
| `NETLIFY` | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| `NS1` | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `OPENSRS` | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| `ORACLE` | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| `OVH` | ❌ | ✅ | ✅ | ❌ | ❔ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ❌ | ✅ | ✅ |
| `PACKETFRAME` | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ❔ |
| `PORKBUN` | ❌ | ✅ | ❌ | ✅ | ❌ | ❔ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |
| `POWERDNS` | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| `ROUTE53` | ✅ | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ | ✅ |
| `RWTH` | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ✅ | ✅ |
| `SOFTLAYER` | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| `TRANSIP` | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❌ | ✅ | ✅ |
| `VULTR` | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |

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
|`HEXONET`|@KaiSchwarz-cnic|
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
code to support this provider, we'd be glad to help in any way.

* [1984 Hosting](https://github.com/StackExchange/dnscontrol/issues/1251) (#1251)
* [Alibaba Cloud DNS](https://github.com/StackExchange/dnscontrol/issues/420)(#420)
* [Constellix (DNSMadeEasy)](https://github.com/StackExchange/dnscontrol/issues/842) (#842)
* [CoreDNS](https://github.com/StackExchange/dnscontrol/issues/1284) (#1284)
* [EnCirca](https://github.com/StackExchange/dnscontrol/issues/1048) (#1048)
* [EU.ORG](https://github.com/StackExchange/dnscontrol/issues/1176) (#1176)
* [Infoblox DNS](https://github.com/StackExchange/dnscontrol/issues/1077) (#1077)
* [Joker.com](https://github.com/StackExchange/dnscontrol/issues/854) (#854)
* [Knot DNS](https://github.com/StackExchange/dnscontrol/issues/436) (#436)
* [RRPPRoxy](https://github.com/StackExchange/dnscontrol/issues/1656) (#1656)
* [RcodeZero](https://github.com/StackExchange/dnscontrol/issues/884) (#884)
* [SynergyWholesale](https://github.com/StackExchange/dnscontrol/issues/1605) (#1605)
