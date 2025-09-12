# Service Providers

## Provider Features

The table below shows various features supported, or not supported by DNSControl providers.
This table is automatically generated from metadata supplied by the provider when they register themselves inside dnscontrol.

A question mark may indicate the feature is not supported by a provider, or it may simply mean
the feature has not been investigated and implemented yet. If a feature you need is missing from
a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.

If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.

Legend:
- ✅ Supported
- ❌ Not supported
- ❓ Not implemented, needs investigation or development
- ❔ Unknown

<!-- provider-matrix-start -->
Jump to a table:

- [Provider Type](#provider-type)
- [Provider API](#provider-api)
- [DNS extensions](#dns-extensions)
- [Service discovery](#service-discovery)
- [Security](#security)
- [DNSSEC](#dnssec)

### Provider Type <!--(table 1/6)-->

| Provider name | Official Support | DNS Provider | Registrar |
| ------------- | ---------------- | ------------ | --------- |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AUTODNS`](autodns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`BIND`](bind.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`CNR`](cnr.md) | <span title="Not supported: Actively maintained provider module.">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DESEC`](desec.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`DYNADOT`](dynadot.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`EASYNAME`](easyname.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GCLOUD`](gcloud.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`GCORE`](gcore.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`HEDNS`](hedns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`HETZNER`](hetzner.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`HEXONET`](hexonet.md) | <span title="Not supported: Actively maintained provider module.">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`INWX`](inwx.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`JOKER`](joker.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`LINODE`](linode.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`LOOPIA`](loopia.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`LUADNS`](luadns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`NETCUP`](netcup.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`NETLIFY`](netlify.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`NS1`](ns1.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`OPENSRS`](opensrs.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`ORACLE`](oracle.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`OVH`](ovh.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`PORKBUN`](porkbun.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`POWERDNS`](powerdns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`ROUTE53`](route53.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`RWTH`](rwth.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`TRANSIP`](transip.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`VULTR`](vultr.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |


### Provider API <!--(table 2/6)-->

| Provider name | [Concurrency Verified](../advanced-features/concurrency-verified.md) | [dual host](../advanced-features/dual-host.md) | create-domains | get-zones |
| ------------- | -------------------------------------------------------------------- | ---------------------------------------------- | -------------- | --------- |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AUTODNS`](autodns.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Supported">✅</span> | <span title="Supported: Azure does not permit modifying the existing NS records, only adding/removing additional records.">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Not implemented">❓</span> | <span title="Supported: Azure does not permit modifying the existing NS records, only adding/removing additional records.">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`BIND`](bind.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported: Driver just maintains list of zone files. It should automatically add missing ones.">✅</span> | <span title="Supported">✅</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Supported">✅</span> | <span title="Not supported: Cloudflare will not work well in situations where it is not the only DNS server">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Supported">✅</span> | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CNR`](cnr.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`DESEC`](desec.md) | <span title="Supported">✅</span> | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Supported">✅</span> | <span title="Not supported: DNSimple does not allow sufficient control over the apex NS records">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Not implemented">❓</span> | <span title="Supported: System NS records cannot be edited. Custom apex NS records can be added/changed/deleted.">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Not implemented">❓</span> | <span title="Not implemented">❓</span> | <span title="Not implemented">❓</span> | <span title="Not implemented">❓</span> |
| [`DYNADOT`](dynadot.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`EASYNAME`](easyname.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Not implemented">❓</span> | <span title="Not supported: Exoscale does not allow sufficient control over the apex NS records">❌</span> | <span title="Not supported">❌</span> | <span title="Not implemented">❓</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported: Can only manage domains registered through their service">❌</span> | <span title="Supported">✅</span> |
| [`GCLOUD`](gcloud.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GCORE`](gcore.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HEDNS`](hedns.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HETZNER`](hetzner.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HEXONET`](hexonet.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not implemented">❓</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`INWX`](inwx.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`JOKER`](joker.md) | <span title="Not supported: Joker API has session-based authentication">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`LINODE`](linode.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`LOOPIA`](loopia.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Not supported: Can only manage domains registered through their service">❌</span> | <span title="Supported">✅</span> |
| [`LUADNS`](luadns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported: Requires domain registered through Web UI">❌</span> | <span title="Supported">✅</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Supported">✅</span> | <span title="Not supported: Doesn&#39;t allow control of apex NS records">❌</span> | <span title="Not supported: Requires domain registered through their service">❌</span> | <span title="Supported">✅</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Not supported: New domains require registration">❌</span> | <span title="Supported">✅</span> |
| [`NETCUP`](netcup.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`NETLIFY`](netlify.md) | <span title="Supported">✅</span> | <span title="Not supported: Netlify does not allow sufficient control over the apex NS records">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`NS1`](ns1.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`OPENSRS`](opensrs.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`ORACLE`](oracle.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`OVH`](ovh.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Not supported: New domains require registration">❌</span> | <span title="Supported">✅</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not implemented">❓</span> |
| [`PORKBUN`](porkbun.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`POWERDNS`](powerdns.md) | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`ROUTE53`](route53.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`RWTH`](rwth.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not implemented">❓</span> |
| [`TRANSIP`](transip.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`VULTR`](vultr.md) | <span title="Not implemented">❓</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |


### DNS extensions <!--(table 3/6)-->

| Provider name | [`ALIAS`](../language-reference/domain-modifiers/ALIAS.md) | [`DNAME`](../language-reference/domain-modifiers/DNAME.md) | [`LOC`](../language-reference/domain-modifiers/LOC.md) | [`PTR`](../language-reference/domain-modifiers/PTR.md) | [`SOA`](../language-reference/domain-modifiers/SOA.md) |
| ------------- | ---------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AUTODNS`](autodns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Not supported: Azure DNS does not provide a generic ALIAS functionality. Use AZURE_ALIAS instead.">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Not supported: Azure DNS does not provide a generic ALIAS functionality. Use AZURE_ALIAS instead.">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`BIND`](bind.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Supported: Bunny flattens CNAME records into A/AAAA records dynamically">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Supported: CF automatically flattens CNAME records into A records dynamically">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`CNR`](cnr.md) | <span title="Supported">✅</span> | <span title="Not supported: Ask for this feature.">❌</span> | <span title="Not supported: Ask for this feature.">❌</span> | <span title="Supported">✅</span> | <span title="Not supported: The SOA record is managed on the DNSZone directly. Data only accessible via StatusDNSZone Request, not via the resource records list. Hard to integrate this into DNSControl by that.">❌</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DESEC`](desec.md) | <span title="Not implemented: Apex aliasing is supported via new SVCB and HTTPS record types. For details, check the deSEC docs.">❓</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Not implemented: Needs custom implementation">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported: According to Domainnameshop this will probably never be supported">❌</span> | <span title="Not supported">❌</span> |
| [`DYNADOT`](dynadot.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EASYNAME`](easyname.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Supported: Only on the bare domain. Otherwise CNAME will be substituted">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`GCLOUD`](gcloud.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`GCORE`](gcore.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported: G-Core supports PTR records only in rDNS zones">✅</span> | <span title="Unknown">❔</span> |
| [`HEDNS`](hedns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`HETZNER`](hetzner.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`HEXONET`](hexonet.md) | <span title="Not supported: Using ALIAS is possible through our extended DNS (X-DNS) service. Feel free to get in touch with us.">❌</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`INWX`](inwx.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> | <span title="Supported: PTR records with empty targets are not supported">✅</span> | <span title="Unknown">❔</span> |
| [`JOKER`](joker.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`LINODE`](linode.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`LOOPIA`](loopia.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported: 💩">❌</span> |
| [`LUADNS`](luadns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported: PTR records are not supported (See Link)">❌</span> | <span title="Unknown">❔</span> |
| [`NETCUP`](netcup.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`NETLIFY`](netlify.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`NS1`](ns1.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`OPENSRS`](opensrs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`ORACLE`](oracle.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`OVH`](ovh.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`PORKBUN`](porkbun.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`POWERDNS`](powerdns.md) | <span title="Supported: Needs to be enabled in PowerDNS first">✅</span> | <span title="Supported: Needs to be enabled in PowerDNS first">✅</span> | <span title="Not implemented: Normalization within the PowerDNS API seems to be buggy, so disabled">❓</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`ROUTE53`](route53.md) | <span title="Not supported: R53 does not provide a generic ALIAS functionality. Use R53_ALIAS instead.">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`RWTH`](rwth.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported: PTR records with empty targets are not supported">✅</span> | <span title="Unknown">❔</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`TRANSIP`](transip.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`VULTR`](vultr.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> |


### Service discovery <!--(table 4/6)-->

| Provider name | [`DHCID`](../language-reference/domain-modifiers/DHCID.md) | [`NAPTR`](../language-reference/domain-modifiers/NAPTR.md) | [`SRV`](../language-reference/domain-modifiers/SRV.md) | [`SVCB`](../language-reference/domain-modifiers/SVCB.md) |
| ------------- | ---------------------------------------------------------- | ---------------------------------------------------------- | ------------------------------------------------------ | -------------------------------------------------------- |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`AUTODNS`](autodns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`BIND`](bind.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`CNR`](cnr.md) | <span title="Not supported: Ask for this feature.">❌</span> | <span title="Supported">✅</span> | <span title="Supported: SRV records with empty targets are not supported">✅</span> | <span title="Not supported: Ask for this feature.">❌</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DESEC`](desec.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Unknown">❔</span> | <span title="Not supported: According to Domainnameshop this will probably never be supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`DYNADOT`](dynadot.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EASYNAME`](easyname.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported: SRV records with empty targets are not supported">✅</span> | <span title="Unknown">❔</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`GCLOUD`](gcloud.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GCORE`](gcore.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported: G-Core doesn&#39;t support SRV records with empty targets">✅</span> | <span title="Supported">✅</span> |
| [`HEDNS`](hedns.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HETZNER`](hetzner.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`HEXONET`](hexonet.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported: SRV records with empty targets are not supported">✅</span> | <span title="Unknown">❔</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`INWX`](inwx.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`JOKER`](joker.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`LINODE`](linode.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`LOOPIA`](loopia.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`LUADNS`](luadns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported: The namecheap web console allows you to make SRV records, but their api does not let you read or set them">❌</span> | <span title="Unknown">❔</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported: SRV records with empty targets are not supported">✅</span> | <span title="Unknown">❔</span> |
| [`NETCUP`](netcup.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`NETLIFY`](netlify.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`NS1`](ns1.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`OPENSRS`](opensrs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`ORACLE`](oracle.md) | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`OVH`](ovh.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`PORKBUN`](porkbun.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`POWERDNS`](powerdns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`ROUTE53`](route53.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`RWTH`](rwth.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported: SRV records with empty targets are not supported.">✅</span> | <span title="Unknown">❔</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |
| [`TRANSIP`](transip.md) | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`VULTR`](vultr.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> |


### Security <!--(table 5/6)-->

| Provider name | [`CAA`](../language-reference/domain-modifiers/CAA.md) | [`HTTPS`](../language-reference/domain-modifiers/HTTPS.md) | [`SSHFP`](../language-reference/domain-modifiers/SSHFP.md) | [`TLSA`](../language-reference/domain-modifiers/TLSA.md) |
| ------------- | ------------------------------------------------------ | ---------------------------------------------------------- | ---------------------------------------------------------- | -------------------------------------------------------- |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AUTODNS`](autodns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Not supported: Azure Private DNS does not support CAA records">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`BIND`](bind.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CNR`](cnr.md) | <span title="Supported">✅</span> | <span title="Not supported: Managed via (Query|Add|Modify|Delete)WebFwd API call. Data not accessible via the resource records list. Hard to integrate this into DNSControl by that.">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DESEC`](desec.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported: Might be supported in the future">❌</span> | <span title="Not implemented: Has support but no documentation. Needs to be investigated.">❓</span> |
| [`DYNADOT`](dynadot.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EASYNAME`](easyname.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GCLOUD`](gcloud.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`GCORE`](gcore.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`HEDNS`](hedns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`HETZNER`](hetzner.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`HEXONET`](hexonet.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`INWX`](inwx.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`JOKER`](joker.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`LINODE`](linode.md) | <span title="Supported: Linode doesn&#39;t support changing the CAA flag">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`LOOPIA`](loopia.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`LUADNS`](luadns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NETCUP`](netcup.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NETLIFY`](netlify.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`NS1`](ns1.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> |
| [`OPENSRS`](opensrs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`ORACLE`](oracle.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`OVH`](ovh.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`PORKBUN`](porkbun.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`POWERDNS`](powerdns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`ROUTE53`](route53.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`RWTH`](rwth.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`TRANSIP`](transip.md) | <span title="Supported">✅</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`VULTR`](vultr.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> | <span title="Not supported">❌</span> |


### DNSSEC <!--(table 6/6)-->

| Provider name | [`AUTODNSSEC`](../language-reference/domain-modifiers/AUTODNSSEC_ON.md) | [`DNSKEY`](../language-reference/domain-modifiers/DNSKEY.md) | [`DS`](../language-reference/domain-modifiers/DS.md) |
| ------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------ | ---------------------------------------------------- |
| [`ADGUARDHOME`](adguardhome.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`AKAMAIEDGEDNS`](akamaiedgedns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`AUTODNS`](autodns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`AXFRDDNS`](axfrddns.md) | <span title="Supported: Just warn when DNSSEC is requested but no RRSIG is found in the AXFR or warn when DNSSEC is not requested but RRSIG are found in the AXFR.">✅</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`AZURE_DNS`](azure_dns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`AZURE_PRIVATE_DNS`](azure_private_dns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`BIND`](bind.md) | <span title="Supported: Just writes out a comment indicating DNSSEC was requested">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`BUNNY_DNS`](bunny_dns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`CLOUDFLAREAPI`](cloudflareapi.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Supported">✅</span> |
| [`CLOUDNS`](cloudns.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`CNR`](cnr.md) | <span title="Not implemented: Ask for this feature.">❓</span> | <span title="Not implemented: Ask for this feature.">❓</span> | <span title="Not implemented: Ask for this feature.">❓</span> |
| [`CSCGLOBAL`](cscglobal.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DESEC`](desec.md) | <span title="Supported: deSEC always signs all records. When trying to disable, a notice is printed.">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`DIGITALOCEAN`](digitalocean.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DNSIMPLE`](dnsimple.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`DNSMADEEASY`](dnsmadeeasy.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`DNSOVERHTTPS`](dnsoverhttps.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`DOMAINNAMESHOP`](domainnameshop.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not implemented">❓</span> |
| [`DYNADOT`](dynadot.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EASYNAME`](easyname.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`EXOSCALE`](exoscale.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`FORTIGATE`](fortigate.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`GANDI_V5`](gandi_v5.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported: Only supports DS records at the apex">❌</span> |
| [`GCLOUD`](gcloud.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`GCORE`](gcore.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`HEDNS`](hedns.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`HETZNER`](hetzner.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> |
| [`HEXONET`](hexonet.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`HOSTINGDE`](hostingde.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> |
| [`HUAWEICLOUD`](huaweicloud.md) | <span title="Not implemented: No public api provided, but can be turned on manually in the console.">❓</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`INTERNETBS`](internetbs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`INWX`](inwx.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not implemented: DS records are only supported at the apex and require a different API call that hasn&#39;t been implemented yet.">❓</span> |
| [`JOKER`](joker.md) | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`LINODE`](linode.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`LOOPIA`](loopia.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported: Only supports DS records at the apex, only for .se and .nu domains; done automatically at back-end.">❌</span> |
| [`LUADNS`](luadns.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`MYTHICBEASTS`](mythicbeasts.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NAMECHEAP`](namecheap.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NAMEDOTCOM`](namedotcom.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NETCUP`](netcup.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`NETLIFY`](netlify.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`NS1`](ns1.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Supported">✅</span> |
| [`OPENSRS`](opensrs.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`ORACLE`](oracle.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`OVH`](ovh.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`PACKETFRAME`](packetframe.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`PORKBUN`](porkbun.md) | <span title="Not supported">❌</span> | <span title="Unknown">❔</span> | <span title="Not supported">❌</span> |
| [`POWERDNS`](powerdns.md) | <span title="Supported">✅</span> | <span title="Supported">✅</span> | <span title="Supported">✅</span> |
| [`REALTIMEREGISTER`](realtimeregister.md) | <span title="Supported">✅</span> | <span title="Unknown">❔</span> | <span title="Not supported: Only for subdomains">❌</span> |
| [`ROUTE53`](route53.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`RWTH`](rwth.md) | <span title="Not implemented: Supported by RWTH but not implemented yet.">❓</span> | <span title="Unknown">❔</span> | <span title="Not implemented: DS records are only supported at the apex and require a different API call that hasn&#39;t been implemented yet.">❓</span> |
| [`SAKURACLOUD`](sakuracloud.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`SOFTLAYER`](softlayer.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |
| [`TRANSIP`](transip.md) | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> | <span title="Not supported">❌</span> |
| [`VULTR`](vultr.md) | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> | <span title="Unknown">❔</span> |

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
|[`AZURE_DNS`](azure_dns.md)|@vatsalyagoel|
|[`BIND`](bind.md)|@tlimoncelli|
|[`CLOUDFLAREAPI`](cloudflareapi.md)|@tresni|
|[`CSCGLOBAL`](cscglobal.md)|@mikenz|
|[`GCLOUD`](gcloud.md)|@riyadhalnur|
|[`ROUTE53`](route53.md)|@tresni|

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
|[`ADGUARDHOME`](adguardhome.md)|@ishanjain28|
|[`AZURE_PRIVATE_DNS`](azure_private_dns.md)|@matthewmgamble|
|[`AKAMAIEDGEDNS`](akamaiedgedns.md)|@edglynes|
|[`AXFRDDNS`](axfrddns.md)|@hnrgrgr|
|[`BUNNY_DNS`](bunny_dns.md)|@ppmathis|
|[`CLOUDFLAREAPI`](cloudflareapi.md)|@tresni|
|[`CLOUDNS`](cloudns.md)|@pragmaton|
|[`CNR`](cnr.md)|@KaiSchwarz-cnic|
|[`CSCGLOBAL`](cscglobal.md)|@Air-New-Zealand|
|[`DESEC`](desec.md)|@D3luxee|
|[`DIGITALOCEAN`](digitalocean.md)|@Deraen|
|[`DNSIMPLE`](dnsimple.md)|@onlyhavecans|
|[`DNSMADEEASY`](dnsmadeeasy.md)|@vojtad|
|[`DNSOVERHTTPS`](dnsoverhttps.md)|@mikenz|
|[`DOMAINNAMESHOP`](domainnameshop.md)|@SimenBai|
|[`EASYNAME`](easyname.md)|@tresni|
|[`EXOSCALE`](exoscale.md)|@pierre-emmanuelJ|
|[`GANDI_V5`](gandi_v5.md)|@TomOnTime|
|[`GCORE`](gcore.md)|@xddxdd|
|[`HEDNS`](hedns.md)|@rblenkinsopp|
|[`HETZNER`](hetzner.md)|@das7pad|
|[`HEXONET`](hexonet.md)|@KaiSchwarz-cnic|
|[`HOSTINGDE`](hostingde.md)|@membero|
|[`HUAWEICLOUD`](huaweicloud.md)|@huihuimoe|
|[`INTERNETBS`](internetbs.md)|@pragmaton|
|[`INWX`](inwx.md)|@patschi|
|[`LINODE`](linode.md)|@koesie10|
|[`LOOPIA`](loopia.md)|@systemcrash|
|[`LUADNS`](luadns.md)|@riku22|
|[`NAMECHEAP`](namecheap.md)|@willpower232|
|[`NETCUP`](netcup.md)|@kordianbruck|
|[`NETLIFY`](netlify.md)|@SphericalKat|
|[`NS1`](ns1.md)|@costasd|
|[`OPENSRS`](opensrs.md)|@philhug|
|[`ORACLE`](oracle.md)|@kallsyms|
|[`OVH`](ovh.md)|@masterzen|
|[`PACKETFRAME`](packetframe.md)|@hamptonmoore|
|[`POWERDNS`](powerdns.md)|@jpbede|
|[`REALTIMEREGISTER`](realtimeregister.md)|@PJEilers|
|[`ROUTE53`](route53.md)|@tresni|
|[`RWTH`](rwth.md)|@MisterErwin|
|[`SAKURACLOUD`](sakuracloud.md)|@ttkzw|
|[`SOFTLAYER`](softlayer.md)|@jamielennox|
|[`TRANSIP`](transip.md)|@blackshadev|
|[`VULTR`](vultr.md)|@pgaskin|

### Requested providers

We have received requests for the following providers. If you would like to contribute
code to support this provider, we'd be glad to help in any way.

*(The list below is sorted alphabetically.)*

* [1984 Hosting](https://github.com/StackExchange/dnscontrol/issues/1251) (#1251)
* [Alibaba Cloud DNS](https://github.com/StackExchange/dnscontrol/issues/420)(#420)
* [BookMyName](https://github.com/StackExchange/dnscontrol/issues/3451) (#3451)
* [Constellix (DNSMadeEasy)](https://github.com/StackExchange/dnscontrol/issues/842) (#842)
* [CoreDNS](https://github.com/StackExchange/dnscontrol/issues/1284) (#1284)
* [EU.ORG](https://github.com/StackExchange/dnscontrol/issues/1176) (#1176)
* [EnCirca](https://github.com/StackExchange/dnscontrol/issues/1048) (#1048)
* [GoDaddy](https://github.com/StackExchange/dnscontrol/issues/2596) (#2596)
* [IPv64](https://github.com/StackExchange/dnscontrol/issues/3471) (#3471)
* [Imperva](https://github.com/StackExchange/dnscontrol/issues/1484) (#1484)
* [Infoblox DNS](https://github.com/StackExchange/dnscontrol/issues/1077) (#1077)
* [Joker.com](https://github.com/StackExchange/dnscontrol/issues/854) (#854)
* [Netim](https://github.com/StackExchange/dnscontrol/issues/3511) (#3511)
* [Plesk](https://github.com/StackExchange/dnscontrol/issues/2261) (#2261)
* [Rackspace Cloud DNS](https://github.com/StackExchange/dnscontrol/issues/2980) (#2980)
* [RcodeZero](https://github.com/StackExchange/dnscontrol/issues/884) (#884)
* [Sav.com](https://github.com/StackExchange/dnscontrol/issues/3633) (#3633)
* [Scaleway](https://github.com/StackExchange/dnscontrol/issues/3606) (#3606)
* [Spaceship](https://github.com/StackExchange/dnscontrol/issues/3452) (#3452)
* [SynergyWholesale](https://github.com/StackExchange/dnscontrol/issues/1605) (#1605)
* [UltraDNS by Neustar / CSCGlobal](https://github.com/StackExchange/dnscontrol/issues/1533) (#1533)
* [Vercel](https://github.com/StackExchange/dnscontrol/issues/3379) (#3379)
* [Yandex Cloud DNS](https://github.com/StackExchange/dnscontrol/issues/3737) (#3737)

#### Q: Why are the above GitHub issues marked "closed"?

A: Following [provider requests](../developer-info/provider-request.md), the request
is closed once it is added to this list. If someone chooses to implement the
provider, they re-open the issue.

#### Q: Would someone write a provider for me?

A: The maintainer of DNSControl does not write new providers.  New providers
are contributed by the community.

DNSControl tries to make writing a provider as easy as possible.  DNSControl
does most of the work for you, you only have to write code to authenticate,
download DNS records, and perform create/modify/delete operations on those
records. Please read the directions for [Writing new DNS
providers](../advanced-features/writing-providers.md).  The DNS maintainers will gladly
coach you through the process.
