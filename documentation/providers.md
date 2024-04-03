# Service Providers

## Provider Features

The table below shows various features supported, or not supported by DNSControl providers.
This table is automatically generated from metadata supplied by the provider when they register themselves inside dnscontrol.

An empty space may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.

<!-- provider-matrix-start -->
| Provider name | Official Support | DNS Provider | Registrar | Concurrency Verified | [`ALIAS`](language-reference/domain/ALIAS.md) | [`CAA`](language-reference/domain/CAA.md) | [`AUTODNSSEC`](language-reference/domain/AUTODNSSEC_ON.md) | [`LOC`](language-reference/domain/LOC.md) | [`NAPTR`](language-reference/domain/NAPTR.md) | [`PTR`](language-reference/domain/PTR.md) | [`SOA`](language-reference/domain/SOA.md) | [`SRV`](language-reference/domain/SRV.md) | [`SSHFP`](language-reference/domain/SSHFP.md) | [`TLSA`](language-reference/domain/TLSA.md) | [`DS`](language-reference/domain/DS.md) | [`DHCID`](language-reference/domain/DHCID.md) | [`DNAME`](language-reference/domain/DNAME.md) | dual host | create-domains | get-zones |
| ------------- | ---------------- | ------------ | --------- | -------------------- | ------------------------------------ | -------------------------------- | ------------------------------------------------- | -------------------------------- | ------------------------------------ | -------------------------------- | -------------------------------- | -------------------------------- | ------------------------------------ | ---------------------------------- | ------------------------------ | ------------------------------------ | ------------------------------------ | --------- | -------------- | --------- |
| [`AKAMAIEDGEDNS`](providers/akamaiedgedns.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`AUTODNS`](providers/autodns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`AXFRDDNS`](providers/axfrddns.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❌ | ❌ | ❌ |
| [`AZURE_DNS`](providers/azure_dns.md) | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`AZURE_PRIVATE_DNS`](providers/azure_private_dns.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`BIND`](providers/bind.md) | ✅ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`BUNNY_DNS`](providers/bunny_dns.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❌ | ✅ | ✅ |
| [`CLOUDFLAREAPI`](providers/cloudflareapi.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ❌ | ✅ | ✅ |
| [`CLOUDNS`](providers/cloudns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ✅ | ✅ |
| [`CSCGLOBAL`](providers/cscglobal.md) | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ |
| [`DESEC`](providers/desec.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ |
| [`DIGITALOCEAN`](providers/digitalocean.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ |
| [`DNSIMPLE`](providers/dnsimple.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ❌ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`DNSMADEEASY`](providers/dnsmadeeasy.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❌ | ❌ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`DNSOVERHTTPS`](providers/dnsoverhttps.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`DOMAINNAMESHOP`](providers/domainnameshop.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| [`DYNADOT`](providers/dynadot.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`EASYNAME`](providers/easyname.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`EXOSCALE`](providers/exoscale.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ |
| [`GANDI_V5`](providers/gandi_v5.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ✅ |
| [`GCLOUD`](providers/gcloud.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`GCORE`](providers/gcore.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❌ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HEDNS`](providers/hedns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HETZNER`](providers/hetzner.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HEXONET`](providers/hexonet.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ❔ |
| [`HOSTINGDE`](providers/hostingde.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`INTERNETBS`](providers/internetbs.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`INWX`](providers/inwx.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`LINODE`](providers/linode.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`LOOPIA`](providers/loopia.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`LUADNS`](providers/luadns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`MSDNS`](providers/msdns.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`MYTHICBEASTS`](providers/mythicbeasts.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`NAMECHEAP`](providers/namecheap.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`NAMEDOTCOM`](providers/namedotcom.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`NETCUP`](providers/netcup.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❌ |
| [`NETLIFY`](providers/netlify.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ✅ | ❌ | ❌ | ❌ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`NS1`](providers/ns1.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`OPENSRS`](providers/opensrs.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`ORACLE`](providers/oracle.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`OVH`](providers/ovh.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`PACKETFRAME`](providers/packetframe.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ |
| [`PORKBUN`](providers/porkbun.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`POWERDNS`](providers/powerdns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`REALTIMEREGISTER`](providers/realtimeregister.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❔ | ❌ | ✅ | ✅ |
| [`ROUTE53`](providers/route53.md) | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`RWTH`](providers/rwth.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`SOFTLAYER`](providers/softlayer.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`TRANSIP`](providers/transip.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ✅ |
| [`VULTR`](providers/vultr.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ |
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

* Maintainers are expected to support their provider and/or help find a new maintainer.
* Maintainers should set up test accounts and periodically verify that all tests pass (`pkg/js/parse_tests` and `integrationTest`).
* Contributors are encouraged to add new tests and refine old ones. (Test-driven development is encouraged.)
* Bugs will be referred to the maintainer or their designate.
* Maintainers must be responsible to bug reports and PRs.  If a maintainer is unresponsive for more than 2 months, we will consider disabling the provider.  First we will put out a call for new maintainer. If nobody volunteers, the provider may be disabled.
* Tom needs to know your real email address.  Please email tlimoncelli at stack over flow dot com so he has it.

Providers in this category and their maintainers are:

|Name|Maintainer|
|---|---|
|[`AZURE_PRIVATE_DNS`](providers/azure_private_dns.md)|@matthewmgamble|
|[`AKAMAIEDGEDNS`](providers/akamaiedgedns.md)|@svernick|
|[`AXFRDDNS`](providers/axfrddns.md)|@hnrgrgr|
|[`BUNNY_DNS`](providers/bunny_dns.md)|@ppmathis|
|[`CLOUDFLAREAPI`](providers/cloudflareapi.md)|@tresni|
|[`CLOUDNS`](providers/cloudns.md)|@pragmaton|
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
|[`INWX`](providers/inwx.md)|@patschi|
|[`LINODE`](providers/linode.md)|@koesie10|
|[`LOOPIA`](providers/loopia.md)|@systemcrash|
|[`LUADNS`](providers/luadns.md)|@riku22|
|[`NAMECHEAP`](providers/namecheap.md)|@willpower232|
|[`NETCUP`](providers/netcup.md)|@kordianbruck|
|[`NETLIFY`](providers/netlify.md)|@SphericalKat|
|[`NS1`](providers/ns1.md)|@costasd|
|[`OPENSRS`](providers/opensrs.md)|@philhug|
|[`ORACLE`](providers/oracle.md)|@kallsyms|
|[`OVH`](providers/ovh.md)|@masterzen|
|[`PACKETFRAME`](providers/packetframe.md)|@hamptonmoore|
|[`POWERDNS`](providers/powerdns.md)|@jpbede|
|[`REALTIMEREGISTER`](providers/realtimeregister.md)|@PJEilers|
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
* [EU.ORG](https://github.com/StackExchange/dnscontrol/issues/1176) (#1176)
* [EnCirca](https://github.com/StackExchange/dnscontrol/issues/1048) (#1048)
* [GoDaddy](https://github.com/StackExchange/dnscontrol/issues/2596) (#2596)
* [Imperva](https://github.com/StackExchange/dnscontrol/issues/1484) (#1484)
* [Infoblox DNS](https://github.com/StackExchange/dnscontrol/issues/1077) (#1077)
* [Joker.com](https://github.com/StackExchange/dnscontrol/issues/854) (#854)
* [Plesk](https://github.com/StackExchange/dnscontrol/issues/2261) (#2261)
* [RRPPRoxy](https://github.com/StackExchange/dnscontrol/issues/1656) (#1656)
* [RcodeZero](https://github.com/StackExchange/dnscontrol/issues/884) (#884)
* [SynergyWholesale](https://github.com/StackExchange/dnscontrol/issues/1605) (#1605)
* [UltraDNS by Neustar / CSCGlobal](https://github.com/StackExchange/dnscontrol/issues/1533) (#1533)

#### Q: Why are the above GitHub issues marked "closed"?

A: Following [the bug triage process](bug-triage.md), the request
is closed once it is added to this list. If someone chooses to implement the
provider, they re-open the issue.

#### Q: Would someone write a provider for me?

A: The maintainer of DNSControl does not write new providers.  New providers
are contributed by the community.

DNSControl tries to make writing a provider as easy as possible.  DNSControl
does most of the work for you, you only have to write code to authenticate,
download DNS records, and perform create/modify/delete operations on those
records. Please read the directions for [Writing new DNS
providers](writing-providers.md).  The DNS maintainers will gladly
coach you through the process.
