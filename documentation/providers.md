# Service Providers

## Provider Features

The table below shows various features supported, or not supported by DNSControl providers.
This table is automatically generated from metadata supplied by the provider when they register themselves inside dnscontrol.

An empty space may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.

<!-- provider-matrix-start -->
| Provider name | Official Support | DNS Provider | Registrar | Concurrency Verified | [`ALIAS`](language-reference/domain-modifiers/ALIAS.md) | [`CAA`](language-reference/domain-modifiers/CAA.md) | [`AUTODNSSEC`](language-reference/domain-modifiers/AUTODNSSEC_ON.md) | [`HTTPS`](language-reference/domain-modifiers/HTTPS.md) | [`LOC`](language-reference/domain-modifiers/LOC.md) | [`NAPTR`](language-reference/domain-modifiers/NAPTR.md) | [`PTR`](language-reference/domain-modifiers/PTR.md) | [`SOA`](language-reference/domain-modifiers/SOA.md) | [`SRV`](language-reference/domain-modifiers/SRV.md) | [`SSHFP`](language-reference/domain-modifiers/SSHFP.md) | [`SVCB`](language-reference/domain-modifiers/SVCB.md) | [`TLSA`](language-reference/domain-modifiers/TLSA.md) | [`DS`](language-reference/domain-modifiers/DS.md) | [`DHCID`](language-reference/domain-modifiers/DHCID.md) | [`DNAME`](language-reference/domain-modifiers/DNAME.md) | [`DNSKEY`](language-reference/domain-modifiers/DNSKEY.md) | dual host | create-domains | get-zones |
| ------------- | ---------------- | ------------ | --------- | -------------------- | ------------------------------------------------------- | --------------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------- | ------------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------- | --------------------------------------------------------- | --------- | -------------- | --------- |
| [`AKAMAIEDGEDNS`](provider/akamaiedgedns.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`AUTODNS`](provider/autodns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❌ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`AXFRDDNS`](provider/axfrddns.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❌ | ❌ | ❌ |
| [`AZURE_DNS`](provider/azure_dns.md) | ✅ | ✅ | ❌ | ✅ | ❌ | ✅ | ❔ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`AZURE_PRIVATE_DNS`](provider/azure_private_dns.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`BIND`](provider/bind.md) | ✅ | ✅ | ❌ | ❌ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| [`BUNNY_DNS`](provider/bunny_dns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ❔ | ❌ | ❌ | ❌ | ❔ | ❔ | ❌ | ✅ | ✅ |
| [`CLOUDFLAREAPI`](provider/cloudflareapi.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ | ✅ |
| [`CLOUDNS`](provider/cloudns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ✅ | ✅ |
| [`CSCGLOBAL`](provider/cscglobal.md) | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ |
| [`DESEC`](provider/desec.md) | ❌ | ✅ | ❌ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ❔ | ✅ | ✅ |
| [`DIGITALOCEAN`](provider/digitalocean.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ |
| [`DNSIMPLE`](provider/dnsimple.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`DNSMADEEASY`](provider/dnsmadeeasy.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❌ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`DNSOVERHTTPS`](provider/dnsoverhttps.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`DOMAINNAMESHOP`](provider/domainnameshop.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❌ | ❔ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ |
| [`DYNADOT`](provider/dynadot.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`EASYNAME`](provider/easyname.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`EXOSCALE`](provider/exoscale.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ |
| [`GANDI_V5`](provider/gandi_v5.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❌ | ✅ |
| [`GCLOUD`](provider/gcloud.md) | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`GCORE`](provider/gcore.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ✅ | ❌ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HEDNS`](provider/hedns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HETZNER`](provider/hetzner.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HEXONET`](provider/hexonet.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ❔ |
| [`HOSTINGDE`](provider/hostingde.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`HUAWEICLOUD`](provider/huaweicloud.md) | ❌ | ✅ | ❌ | ❔ | ❌ | ✅ | ❔ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`INTERNETBS`](provider/internetbs.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`INWX`](provider/inwx.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`LINODE`](provider/linode.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`LOOPIA`](provider/loopia.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❌ | ❔ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`LUADNS`](provider/luadns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`MSDNS`](provider/msdns.md) | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❔ | ❔ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`MYTHICBEASTS`](provider/mythicbeasts.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`NAMECHEAP`](provider/namecheap.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ❌ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`NAMEDOTCOM`](provider/namedotcom.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`NETCUP`](provider/netcup.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ✅ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❌ |
| [`NETLIFY`](provider/netlify.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ❔ | ❌ | ❌ | ❌ | ❔ | ✅ | ❌ | ❔ | ❌ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`NS1`](provider/ns1.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ❔ | ✅ | ❔ | ✅ | ✅ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ |
| [`OPENSRS`](provider/opensrs.md) | ❌ | ❌ | ✅ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`ORACLE`](provider/oracle.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`OVH`](provider/ovh.md) | ❌ | ✅ | ✅ | ❌ | ❌ | ✅ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ✅ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ✅ | ❌ | ✅ |
| [`PACKETFRAME`](provider/packetframe.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ❔ |
| [`PORKBUN`](provider/porkbun.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ❔ | ❌ | ❔ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ | ❔ | ✅ | ❌ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`POWERDNS`](provider/powerdns.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ❔ | ✅ | ✅ | ✅ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`REALTIMEREGISTER`](provider/realtimeregister.md) | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❔ | ✅ | ✅ | ❌ | ❌ | ✅ | ✅ | ❔ | ✅ | ❌ | ❌ | ❔ | ❔ | ❌ | ✅ | ✅ |
| [`ROUTE53`](provider/route53.md) | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ❔ | ❔ | ❌ | ❔ | ✅ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ | ✅ |
| [`RWTH`](provider/rwth.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❔ | ❌ | ❌ | ✅ | ❔ | ✅ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❌ | ❌ | ✅ |
| [`SAKURACLOUD`](provider/sakuracloud.md) | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ |
| [`SOFTLAYER`](provider/softlayer.md) | ❌ | ✅ | ❌ | ❌ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ | ❔ | ❔ | ✅ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❔ | ❌ | ❔ |
| [`TRANSIP`](provider/transip.md) | ❌ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |
| [`VULTR`](provider/vultr.md) | ❌ | ✅ | ❌ | ❌ | ❌ | ✅ | ❔ | ❔ | ❌ | ❔ | ❌ | ❔ | ✅ | ✅ | ❔ | ❌ | ❔ | ❔ | ❔ | ❔ | ❔ | ✅ | ✅ |
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
|[`AZURE_DNS`](provider/azure_dns.md)|@vatsalyagoel|
|[`BIND`](provider/bind.md)|@tlimoncelli|
|[`CLOUDFLAREAPI`](provider/cloudflareapi.md)|@tresni|
|[`CSCGLOBAL`](provider/cscglobal.md)|@mikenz|
|[`GCLOUD`](provider/gcloud.md)|@riyadhalnur|
|[`MSDNS`](provider/msdns.md)|@tlimoncelli|
|[`ROUTE53`](provider/route53.md)|@tresni|

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
|[`AZURE_PRIVATE_DNS`](provider/azure_private_dns.md)|@matthewmgamble|
|[`AKAMAIEDGEDNS`](provider/akamaiedgedns.md)|@edglynes|
|[`AXFRDDNS`](provider/axfrddns.md)|@hnrgrgr|
|[`BUNNY_DNS`](provider/bunny_dns.md)|@ppmathis|
|[`CLOUDFLAREAPI`](provider/cloudflareapi.md)|@tresni|
|[`CLOUDNS`](provider/cloudns.md)|@pragmaton|
|[`CSCGLOBAL`](provider/cscglobal.md)|@Air-New-Zealand|
|[`DESEC`](provider/desec.md)|@D3luxee|
|[`DIGITALOCEAN`](provider/digitalocean.md)|@Deraen|
|[`DNSIMPLE`](provider/dnsimple.md)|@onlyhavecans|
|[`DNSMADEEASY`](provider/dnsmadeeasy.md)|@vojtad|
|[`DNSOVERHTTPS`](provider/dnsoverhttps.md)|@mikenz|
|[`DOMAINNAMESHOP`](provider/domainnameshop.md)|@SimenBai|
|[`EASYNAME`](provider/easyname.md)|@tresni|
|[`EXOSCALE`](provider/exoscale.md)|@pierre-emmanuelJ|
|[`GANDI_V5`](provider/gandi_v5.md)|@TomOnTime|
|[`GCORE`](provider/gcore.md)|@xddxdd|
|[`HEDNS`](provider/hedns.md)|@rblenkinsopp|
|[`HETZNER`](provider/hetzner.md)|@das7pad|
|[`HEXONET`](provider/hexonet.md)|@KaiSchwarz-cnic|
|[`HOSTINGDE`](provider/hostingde.md)|@membero|
|[`HUAWEICLOUD`](provider/huaweicloud.md)|@huihuimoe|
|[`INTERNETBS`](provider/internetbs.md)|@pragmaton|
|[`INWX`](provider/inwx.md)|@patschi|
|[`LINODE`](provider/linode.md)|@koesie10|
|[`LOOPIA`](provider/loopia.md)|@systemcrash|
|[`LUADNS`](provider/luadns.md)|@riku22|
|[`NAMECHEAP`](provider/namecheap.md)|@willpower232|
|[`NETCUP`](provider/netcup.md)|@kordianbruck|
|[`NETLIFY`](provider/netlify.md)|@SphericalKat|
|[`NS1`](provider/ns1.md)|@costasd|
|[`OPENSRS`](provider/opensrs.md)|@philhug|
|[`ORACLE`](provider/oracle.md)|@kallsyms|
|[`OVH`](provider/ovh.md)|@masterzen|
|[`PACKETFRAME`](provider/packetframe.md)|@hamptonmoore|
|[`POWERDNS`](provider/powerdns.md)|@jpbede|
|[`REALTIMEREGISTER`](provider/realtimeregister.md)|@PJEilers|
|[`ROUTE53`](provider/route53.md)|@tresni|
|[`RWTH`](provider/rwth.md)|@MisterErwin|
|[`SAKURACLOUD`](provider/sakuracloud.md)|@ttkzw|
|[`SOFTLAYER`](provider/softlayer.md)|@jamielennox|
|[`TRANSIP`](provider/transip.md)|@blackshadev|
|[`VULTR`](provider/vultr.md)|@pgaskin|

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
