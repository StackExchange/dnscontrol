# Service Providers

## Provider Features

The table below shows various features supported, or not supported by DNSControl providers.
This table is automatically generated from metadata supplied by the provider when they register themselves inside dnscontrol.

An empty space may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.

<!-- provider-matrix-start -->
| Provider name | Official Support | DNS Provider | Registrar | [`ALIAS`](../language_reference/domain_modifier_functions/ALIAS.md) | [`CAA`](../language_reference/domain_modifier_functions/CAA.md) | [`AUTODNSSEC`](../language_reference/domain_modifier_functions/AUTODNSSEC_ON.md) | [`LOC`](../language_reference/domain_modifier_functions/LOC.md) | [`NAPTR`](../language_reference/domain_modifier_functions/NAPTR.md) | [`PTR`](../language_reference/domain_modifier_functions/PTR.md) | [`SOA`](../language_reference/domain_modifier_functions/SOA.md) | [`SRV`](../language_reference/domain_modifier_functions/SRV.md) | [`SSHFP`](../language_reference/domain_modifier_functions/SSHFP.md) | [`TLSA`](../language_reference/domain_modifier_functions/TLSA.md) | [`DS`](../language_reference/domain_modifier_functions/DS.md) | dual host | create-domains | [`NO_PURGE`](../language_reference/domain_modifier_functions/NO_PURGE.md) | get-zones |
| ------------- | ---------------- | ------------ | --------- | ------------------------------------------------------------------- | --------------------------------------------------------------- | -------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------------- | ------------------------------------------------------------- | --------- | -------------- | ------------------------------------------------------------------------- | --------- |
| [`AKAMAIEDGEDNS`](service_providers/akamaiedgedns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ |
| [`AUTODNS`](service_providers/autodns.md) | ❌ | ✅ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| [`AXFRDDNS`](service_providers/axfrddns.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❌ | ❌ | ❌ | ❌ |
| [`AZURE_DNS`](service_providers/azure_dns.md) | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ |
| [`BIND`](service_providers/bind.md) | ✅ | ✅ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ |
| [`CLOUDFLAREAPI`](service_providers/cloudflareapi.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❌ | ✅ | ✅ | ✅ |
| [`CLOUDNS`](service_providers/cloudns.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`CSCGLOBAL`](service_providers/cscglobal.md) | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ✅ |
| [`DESEC`](service_providers/desec.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ |
| [`DIGITALOCEAN`](service_providers/digitalocean.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`DNSIMPLE`](service_providers/dnsimple.md) | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| [`DNSMADEEASY`](service_providers/dnsmadeeasy.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| [`DNSOVERHTTPS`](service_providers/dnsoverhttps.md) | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| [`DOMAINNAMESHOP`](service_providers/domainnameshop.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ |
| [`EASYNAME`](service_providers/easyname.md) | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| [`EXOSCALE`](service_providers/exoscale.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❌ | ❔ | ❌ | ❌ | ✅ | ❔ |
| [`GANDI_V5`](service_providers/gandi_v5.md) | ❌ | ✅ | ✅ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ✅ |
| [`GCLOUD`](service_providers/gcloud.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ |
| [`GCORE`](service_providers/gcore.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| [`HEDNS`](service_providers/hedns.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ |
| [`HETZNER`](service_providers/hetzner.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`HEXONET`](service_providers/hexonet.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ |
| [`HOSTINGDE`](service_providers/hostingde.md) | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`INTERNETBS`](service_providers/internetbs.md) | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| [`INWX`](service_providers/inwx.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ |
| [`LINODE`](service_providers/linode.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ✅ |
| [`LOOPIA`](service_providers/loopia.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ✅ |
| [`LUADNS`](service_providers/luadns.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ |
| [`MSDNS`](service_providers/msdns.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ✅ |
| [`NAMECHEAP`](service_providers/namecheap.md) | ❌ | ✅ | ✅ | ✅ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❌ | ❌ | ✅ |
| [`NAMEDOTCOM`](service_providers/namedotcom.md) | ❌ | ✅ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ | ✅ |
| [`NETCUP`](service_providers/netcup.md) | ❌ | ✅ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ❌ |
| [`NETLIFY`](service_providers/netlify.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| [`NS1`](service_providers/ns1.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`OPENSRS`](service_providers/opensrs.md) | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| [`ORACLE`](service_providers/oracle.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ |
| [`OVH`](service_providers/ovh.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ❌ | ✅ | ✅ |
| [`PACKETFRAME`](service_providers/packetframe.md) | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ❔ |
| [`PORKBUN`](service_providers/porkbun.md) | ❌ | ✅ | ❌ | ✅ | ❔ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ |
| [`POWERDNS`](service_providers/powerdns.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`ROUTE53`](service_providers/route53.md) | ✅ | ✅ | ✅ | ❌ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ | ✅ |
| [`RWTH`](service_providers/rwth.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ✅ | ✅ |
| [`SOFTLAYER`](service_providers/softlayer.md) | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ | ❔ |
| [`TRANSIP`](service_providers/transip.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❌ | ✅ | ✅ |
| [`VULTR`](service_providers/vultr.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
<!-- provider-matrix-end -->

### Providers with "official support"

Official support means:

* New releases will block if any of these providers do not pass integration tests.
* The DNSControl maintainers prioritize fixing bugs in these providers (though we gladly accept PRs).
* New features will work on these providers (unless the provider does not support it).
* StackOverflow maintains test accounts with those providers for running integration tests.

Providers in this category and their maintainers are:

|Name|Maintainer|
|---|---|
|[`AZURE_DNS`](providers/azure_dns.md)|@vatsalyagoel|
|[`BIND`](providers/bind.md)|@tlimoncelli|
|[`CLOUDFLAREAPI`](providers/cloudflareapi.md)|@tresni|
|[`CSCGLOBAL`](providers/cscglobal.md)|@mikenz|
|[`GCLOUD`](providers/gcloud.md)|@riyadhalnur|
|[`MSDNS`](providers/msdns.md)|@tlimoncelli|
|[`ROUTE53`](providers/route53.md)|@tresni|

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
|[`AKAMAIEDGEDNS`](providers/akamaiedgedns.md)|@svernick|
|[`AXFRDDNS`](providers/axfrddns.md)|@hnrgrgr|
|[`CLOUDFLAREAPI`](providers/cloudflareapi.md)|@tresni|
|[`CLOUDNS`](providers/CLOUDNS.md)|@pragmaton|
|[`CSCGLOBAL`](providers/cscglobal.md)|@Air-New-Zealand|
|[`DESEC`](providers/desec.md)|@D3luxee|
|[`DIGITALOCEAN`](providers/digitalocean.md)|@Deraen|
|[`DNSIMPLE`](providers/dnsimple.md)|@onlyhavecans|
|[`DNSMADEEASY`](providers/dnsmadeeasy.md)|@vojtad|
|[`DNSOVERHTTPS`](providers/dnsoverhttps.md)|@mikenz|
|[`DOMAINNAMESHOP`](providers/domainnameshop.md)|@SimenBai|
|[`EASYNAME`](providers/easyname.md)|@tresni|
|[`EXOSCALE`](providers/exoscale.md)|@pierre-emmanuelJ|
|[`GANDI_V5`](providers/gandi_v5.md)|@TomOnTime|
|[`GCORE`](providers/gcore.md)|@xddxdd|
|[`HEDNS`](providers/hedns.md)|@rblenkinsopp|
|[`HETZNER`](providers/hetzner.md)|@das7pad|
|[`HEXONET`](providers/hexonet.md)|@KaiSchwarz-cnic|
|[`HOSTINGDE`](providers/hostingde.md)|@membero|
|[`INTERNETBS`](providers/internetbs.md)|@pragmaton|
|[`INWX`](providers/inwx.md)|@svenpeter42|
|[`LINODE`](providers/linode.md)|@koesie10|
|[`LOOPIA`](providers/loopia.md)|@systemcrash|
|[`LUADNS`](providers/luadns.md)|@riku22|
|[`NAMECHEAP`](providers/namecheap.md)|@willpower232|
|[`NETCUP`](providers/netcup.md)|@kordianbruck|
|[`NETLIFY`](providers/netlify.md)|@SphericalKat|
|[`NS1`](providers/ns1.md)|@costasd|
|[`OPENSRS`](providers/opensrs.md)|@pierre-emmanuelJ|
|[`ORACLE`](providers/oracle.md)|@kallsyms|
|[`OVH`](providers/ovh.md)|@masterzen|
|[`PACKETFRAME`](providers/packetframe.md)|@hamptonmoore|
|[`POWERDNS`](providers/powerdns.md)|@jpbede|
|[`ROUTE53`](providers/route53.md)|@tresni|
|[`RWTH`](providers/rwth.md)|@MisterErwin|
|[`SOFTLAYER`](providers/softlayer.md)|@jamielennox|
|[`TRANSIP`](providers/transip.md)|@blackshadev|
|[`VULTR`](providers/vultr.md)|@pgaskin|

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
* [RRPPRoxy](https://github.com/StackExchange/dnscontrol/issues/1656) (#1656)
* [RcodeZero](https://github.com/StackExchange/dnscontrol/issues/884) (#884)
* [SynergyWholesale](https://github.com/StackExchange/dnscontrol/issues/1605) (#1605)
