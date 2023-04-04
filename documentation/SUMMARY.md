# Table of contents

* [Introduction to DNSControl](index.md)

## Getting Started

* [Overview](getting-started.md)
* [Examples](examples.md)
* [Migrating zones to DNSControl](migrating.md)
* [TypeScript autocomplete and type checking](typescript.md)
<!-- LANG_REF start -->

## Language Reference
  * [JavaScript DSL](language_reference/JavaScript_DSL.md)
  * Domain Modifier Functions
    * [A](language_reference/domain_modifier_functions/A.md)
    * [AAAA](language_reference/domain_modifier_functions/AAAA.md)
    * [ALIAS](language_reference/domain_modifier_functions/ALIAS.md)
    * [AUTODNSSEC_OFF](language_reference/domain_modifier_functions/AUTODNSSEC_OFF.md)
    * [AUTODNSSEC_ON](language_reference/domain_modifier_functions/AUTODNSSEC_ON.md)
    * [CAA](language_reference/domain_modifier_functions/CAA.md)
    * [CNAME](language_reference/domain_modifier_functions/CNAME.md)
    * [DefaultTTL](language_reference/domain_modifier_functions/DefaultTTL.md)
    * [DnsProvider](language_reference/domain_modifier_functions/DnsProvider.md)
    * [DS](language_reference/domain_modifier_functions/DS.md)
    * [FRAME](language_reference/domain_modifier_functions/FRAME.md)
    * [IGNORE](language_reference/domain_modifier_functions/IGNORE.md)
    * [IGNORE_NAME](language_reference/domain_modifier_functions/IGNORE_NAME.md)
    * [IGNORE_TARGET](language_reference/domain_modifier_functions/IGNORE_TARGET.md)
    * [IMPORT_TRANSFORM](language_reference/domain_modifier_functions/IMPORT_TRANSFORM.md)
    * [INCLUDE](language_reference/domain_modifier_functions/INCLUDE.md)
    * [LOC](language_reference/domain_modifier_functions/LOC.md)
    * [MX](language_reference/domain_modifier_functions/MX.md)
    * [NAMESERVER](language_reference/domain_modifier_functions/NAMESERVER.md)
    * [NAMESERVER_TTL](language_reference/domain_modifier_functions/NAMESERVER_TTL.md)
    * [NAPTR](language_reference/domain_modifier_functions/NAPTR.md)
    * [NO_PURGE](language_reference/domain_modifier_functions/NO_PURGE.md)
    * [NS](language_reference/domain_modifier_functions/NS.md)
    * [PTR](language_reference/domain_modifier_functions/PTR.md)
    * [PURGE](language_reference/domain_modifier_functions/PURGE.md)
    * [SOA](language_reference/domain_modifier_functions/SOA.md)
    * [SRV](language_reference/domain_modifier_functions/SRV.md)
    * [SSHFP](language_reference/domain_modifier_functions/SSHFP.md)
    * [TLSA](language_reference/domain_modifier_functions/TLSA.md)
    * [TXT](language_reference/domain_modifier_functions/TXT.md)
    * [URL](language_reference/domain_modifier_functions/URL.md)
    * [URL301](language_reference/domain_modifier_functions/URL301.md)
    * Service Provider Specific
      * Akamai Edge Dns
        * [AKAMAICDN](language_reference/domain_modifier_functions/service_provider_specific/akamai_edge_dns/AKAMAICDN.md)
      * Amazon Route 53
        * [R53_ALIAS](language_reference/domain_modifier_functions/service_provider_specific/amazon_route_53/R53_ALIAS.md)
      * Azure Dns
        * [AZURE_ALIAS](language_reference/domain_modifier_functions/service_provider_specific/azure_dns/AZURE_ALIAS.md)
      * Cloudflare Dns
        * [CF_REDIRECT](language_reference/domain_modifier_functions/service_provider_specific/cloudflare_dns/CF_REDIRECT.md)
        * [CF_TEMP_REDIRECT](language_reference/domain_modifier_functions/service_provider_specific/cloudflare_dns/CF_TEMP_REDIRECT.md)
        * [CF_WORKER_ROUTE](language_reference/domain_modifier_functions/service_provider_specific/cloudflare_dns/CF_WORKER_ROUTE.md)
      * ClouDNS
        * [CLOUDNS_WR](language_reference/domain_modifier_functions/service_provider_specific/ClouDNS/CLOUDNS_WR.md)
      * NS1
        * [NS1_URLFWD](language_reference/domain_modifier_functions/service_provider_specific/NS1/NS1_URLFWD.md)
  * Record Modifier Functions
    * [CAA_BUILDER](language_reference/record_modifier_functions/CAA_BUILDER.md)
    * [DMARC_BUILDER](language_reference/record_modifier_functions/DMARC_BUILDER.md)
    * [LOC_BUILDER_DD](language_reference/record_modifier_functions/LOC_BUILDER_DD.md)
    * [LOC_BUILDER_DMM_STR](language_reference/record_modifier_functions/LOC_BUILDER_DMM_STR.md)
    * [LOC_BUILDER_DMS_STR](language_reference/record_modifier_functions/LOC_BUILDER_DMS_STR.md)
    * [LOC_BUILDER_STR](language_reference/record_modifier_functions/LOC_BUILDER_STR.md)
    * [SPF_BUILDER](language_reference/record_modifier_functions/SPF_BUILDER.md)
    * [TTL](language_reference/record_modifier_functions/TTL.md)
    * Service Provider Specific
      * Amazon Route 53
        * [R53_ZONE](language_reference/record_modifier_functions/service_provider_specific/amazon_route_53/R53_ZONE.md)
  * Top Level Functions
    * [D](language_reference/top_level_functions/D.md)
    * [D_EXTEND](language_reference/top_level_functions/D_EXTEND.md)
    * [DEFAULTS](language_reference/top_level_functions/DEFAULTS.md)
    * [DOMAIN_ELSEWHERE](language_reference/top_level_functions/DOMAIN_ELSEWHERE.md)
    * [DOMAIN_ELSEWHERE_AUTO](language_reference/top_level_functions/DOMAIN_ELSEWHERE_AUTO.md)
    * [FETCH](language_reference/top_level_functions/FETCH.md)
    * [getConfiguredDomains](language_reference/top_level_functions/getConfiguredDomains.md)
    * [IP](language_reference/top_level_functions/IP.md)
    * [NewDnsProvider](language_reference/top_level_functions/NewDnsProvider.md)
    * [NewRegistrar](language_reference/top_level_functions/NewRegistrar.md)
    * [PANIC](language_reference/top_level_functions/PANIC.md)
    * [require](language_reference/top_level_functions/require.md)
    * [require_glob](language_reference/top_level_functions/require_glob.md)
    * [REV](language_reference/top_level_functions/REV.md)
<!-- LANG_REF end -->
* [Why CNAME/MX/NS targets require a "dot"](why-the-dot.md)
<!-- PROVIDER start -->

## Service Providers
  * [Providers](service_providers/providers.md)
    * [Autodns](service_providers/providers/autodns.md)
    * [Axfrddns](service_providers/providers/axfrddns.md)
    * [Azure Dns](service_providers/providers/azure_dns.md)
    * [Bind](service_providers/providers/bind.md)
    * [Cloudflareapi](service_providers/providers/cloudflareapi.md)
    * [Cloudns](service_providers/providers/cloudns.md)
    * [Cscglobal](service_providers/providers/cscglobal.md)
    * [Desec](service_providers/providers/desec.md)
    * [Digitalocean](service_providers/providers/digitalocean.md)
    * [Dnsimple](service_providers/providers/dnsimple.md)
    * [Dnsmadeeasy](service_providers/providers/dnsmadeeasy.md)
    * [Dnsoverhttps](service_providers/providers/dnsoverhttps.md)
    * [Domainnameshop](service_providers/providers/domainnameshop.md)
    * [Easyname](service_providers/providers/easyname.md)
    * [Gandi V5](service_providers/providers/gandi_v5.md)
    * [Gcloud](service_providers/providers/gcloud.md)
    * [Gcore](service_providers/providers/gcore.md)
    * [Hedns](service_providers/providers/hedns.md)
    * [Hetzner](service_providers/providers/hetzner.md)
    * [Hexonet](service_providers/providers/hexonet.md)
    * [Hostingde](service_providers/providers/hostingde.md)
    * [Internetbs](service_providers/providers/internetbs.md)
    * [Inwx](service_providers/providers/inwx.md)
    * [Linode](service_providers/providers/linode.md)
    * [Loopia](service_providers/providers/loopia.md)
    * [Luadns](service_providers/providers/luadns.md)
    * [Msdns](service_providers/providers/msdns.md)
    * [Namecheap](service_providers/providers/namecheap.md)
    * [Namedotcom](service_providers/providers/namedotcom.md)
    * [Netcup](service_providers/providers/netcup.md)
    * [Netlify](service_providers/providers/netlify.md)
    * [Ns1](service_providers/providers/ns1.md)
    * [Oracle](service_providers/providers/oracle.md)
    * [Ovh](service_providers/providers/ovh.md)
    * [Packetframe](service_providers/providers/packetframe.md)
    * [Porkbun](service_providers/providers/porkbun.md)
    * [Powerdns](service_providers/providers/powerdns.md)
    * [Route53](service_providers/providers/route53.md)
    * [Rwth](service_providers/providers/rwth.md)
    * [Softlayer](service_providers/providers/softlayer.md)
    * [Transip](service_providers/providers/transip.md)
    * [Vultr](service_providers/providers/vultr.md)
<!-- PROVIDER end -->
## Commands

* [creds.json](creds-json.md)
* [check-creds](check-creds.md)
* [get-certs](get-certs.md)
* [get-zones](get-zones.md)

## Advanced features

* [CI/CD example for GitLab](ci-cd-gitlab.md)
* [CLI variables](cli-variables.md)
* [Nameservers and Delegations](nameservers.md)
* [Notifications](notifications.md)
* [Useful code tricks](code-tricks.md)

## Developer info

* [Style Guide](styleguide.md)
* [DNSControl is an opinionated system](opinions.md)
* [Writing new DNS providers](writing-providers.md)
* [Creating new DNS Resource Types (rtypes)](adding-new-rtypes.md)
* [Integration Tests](integration-tests.md)
* [Unit Testing DNS Data](unittests.md)
* [Bug Triage Process](bug-triage.md)
* [Bring-Your-Own-Secrets for automated testing](byo-secrets.md)
* [Debugging with dlv](debugging-with-dlv.md)
* [ALIAS Records](alias.md)
* [TXT record testing](testing-txt-records.md)

## Release

* [How to build and ship a release](release-engineering.md)
* [Changelog v3.16.0](v316.md)
* [GitHub releases](https://github.com/StackExchange/dnscontrol/releases/latest)
