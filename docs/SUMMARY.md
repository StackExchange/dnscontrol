# Table of contents

* [Introduction to DNSControl](index.md)

## Getting Started

* [Overview](getting-started.md)
* [Examples](examples.md)
* [Migrating zones to DNSControl](migrating.md)
* [TypeScript autocomplete and type checking](typescript.md)

## Language Reference

* [JavaScript DSL](js.md)
* Top Level Functions
  * [D](functions/global/D.md)
  * [DEFAULTS](functions/global/DEFAULTS.md)
  * [DOMAIN_ELSEWHERE](functions/global/DOMAIN_ELSEWHERE.md)
  * [DOMAIN_ELSEWHERE_AUTO](functions/global/DOMAIN_ELSEWHERE_AUTO.md)
  * [D_EXTEND](functions/global/D_EXTEND.md)
  * [FETCH](functions/global/FETCH.md)
  * [IP](functions/global/IP.md)
  * [NewDnsProvider](functions/global/NewDnsProvider.md)
  * [NewRegistrar](functions/global/NewRegistrar.md)
  * [PANIC](functions/global/PANIC.md)
  * [REV](functions/global/REV.md)
  * [getConfiguredDomains](functions/global/getConfiguredDomains.md)
  * [require](functions/global/require.md)
  * [require_glob](functions/global/require_glob.md)
* Domain Modifiers
    * [A](functions/domain/A.md)
    * [AAAA](functions/domain/AAAA.md)
    * [ALIAS](functions/domain/ALIAS.md)
    * [AUTODNSSEC_OFF](functions/domain/AUTODNSSEC_OFF.md)
    * [AUTODNSSEC_ON](functions/domain/AUTODNSSEC_ON.md)
    * [CAA](functions/domain/CAA.md)
    * [CNAME](functions/domain/CNAME.md)
    * [DS](functions/domain/DS.md)
    * [DefaultTTL](functions/domain/DefaultTTL.md)
    * [DnsProvider](functions/domain/DnsProvider.md)
    * [FRAME](functions/domain/FRAME.md)
    * [IGNORE](functions/domain/IGNORE.md)
    * [IGNORE_NAME](functions/domain/IGNORE_NAME.md)
    * [IGNORE_TARGET](functions/domain/IGNORE_TARGET.md)
    * [IMPORT_TRANSFORM](functions/domain/IMPORT_TRANSFORM.md)
    * [INCLUDE](functions/domain/INCLUDE.md)
    * [MX](functions/domain/MX.md)
    * [NAMESERVER](functions/domain/NAMESERVER.md)
    * [NAMESERVER_TTL](functions/domain/NAMESERVER_TTL.md)
    * [NO_PURGE](functions/domain/NO_PURGE.md)
    * [NS](functions/domain/NS.md)
    * [PTR](functions/domain/PTR.md)
    * [PURGE](functions/domain/PURGE.md)
    * [SOA](functions/domain/SOA.md)
    * [SRV](functions/domain/SRV.md)
    * [SSHFP](functions/domain/SSHFP.md)
    * [TLSA](functions/domain/TLSA.md)
    * [TXT](functions/domain/TXT.md)
    * [URL](functions/domain/URL.md)
    * [URL301](functions/domain/URL301.md)
    * Service Provider specific
        * Akamai Edge Dns
            * [AKAMAICDN](functions/domain/AKAMAICDN.md)
        * Amazon Route 53
            * [R53_ALIAS](functions/domain/R53_ALIAS.md)
        * Azure DNS
            * [AZURE_ALIAS](functions/domain/AZURE_ALIAS.md)
        * Cloudflare DNS
            * [CF_REDIRECT](functions/domain/CF_REDIRECT.md)
            * [CF_TEMP_REDIRECT](functions/domain/CF_TEMP_REDIRECT.md)
            * [CF_WORKER_ROUTE](functions/domain/CF_WORKER_ROUTE.md)
        * ClouDNS
            * [CLOUDNS_WR](functions/domain/CLOUDNS_WR.md)
        * NS1
            * [NS1_URLFWD](functions/domain/NS1_URLFWD.md)
* Record Modifiers
    * [CAA_BUILDER](functions/record/CAA_BUILDER.md)
    * [DMARC_BUILDER](functions/record/DMARC_BUILDER.md)
    * [SPF_BUILDER](functions/record/SPF_BUILDER.md)
    * [TTL](functions/record/TTL.md)
    * Service Provider specific
        * Amazon Route 53
            * [R53_ZONE](functions/record/R53_ZONE.md)
* [Why CNAME/MX/NS targets require a "dot"](why-the-dot.md)

## Service Providers

* [Providers](providers.md)
    * [Akamai Edge DNS](_providers/akamaiedgedns.md)
    * [AutoDNS](_providers/autodns.md)
    * [AXFR+DDNS](_providers/axfrddns.md)
    * [Azure DNS](_providers/azure_dns.md)
    * [BIND](_providers/bind.md)
    * [Cloudflare](_providers/cloudflareapi.md)
    * [ClouDNS](_providers/cloudns.md)
    * [CSC Global](_providers/cscglobal.md)
    * [deSEC](_providers/desec.md)
    * [DigitalOcean](_providers/digitalocean.md)
    * [DNSimple](_providers/dnsimple.md)
    * [DNS Made Simple](_providers/dnsmadeeasy.md)
    * [DNS-over-HTTPS](_providers/dnsoverhttps.md)
    * [DOMAINNAMESHOP](_providers/domainnameshop.md)
    * [easyname](_providers/easyname.md)
    * [Gandi_v5](_providers/gandi_v5.md)
    * [Google Cloud DNS](_providers/gcloud.md)
    * [Gcore](_providers/gcore.md)
    * [Hurricane Electric DNS](_providers/hedns.md)
    * [Hetzner DNS Console](_providers/hetzner.md)
    * [HEXONET](_providers/hexonet.md)
    * [hosting.de](_providers/hostingde.md)
    * [Internet.bs](_providers/internetbs.md)
    * [INWX](_providers/inwx.md)
    * [Linode](_providers/linode.md)
    * [Microsoft DNS Server on Microsoft Windows Server](_providers/msdns.md)
    * [Namecheap](_providers/namecheap.md)
    * [Name.com](_providers/namedotcom.md)
    * [Netcup](_providers/netcup.md)
    * [Netlify](_providers/netlify.md)
    * [NS1](_providers/ns1.md)
    * [Oracle Cloud](_providers/oracle.md)
    * [OVH](_providers/ovh.md)
    * [Packetframe](_providers/packetframe.md)
    * [Porkbun](_providers/porkbun.md)
    * [PowerDNS](_providers/powerdns.md)
    * [Amazon Route 53](_providers/route53.md)
    * [RWTH DNS-Admin](_providers/rwth.md)
    * [SoftLayer DNS](_providers/softlayer.md)
    * [TransIP DNS](_providers/transip.md)
    * [Vultr](_providers/vultr.md)

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

* [ALIAS Records](alias.md)
* [Bug Triage Process](bug-triage.md)
* [Bring-Your-Own-Secrets for automated testing](byo-secrets.md)
* [Writing new DNS providers](writing-providers.md)
* [Creating new DNS Resource Types (rtypes)](adding-new-rtypes.md)
* [TXT record testing](testing-txt-records.md)
* [Unit Testing DNS Data](unittests.md)
* [DNSControl is an opinionated system](opinions.md)

## Release

* [How to build and ship a release](release-engineering.md)
* [Changelog v3.16.0](v316.md)
* [GitHub releases](https://github.com/StackExchange/dnscontrol/releases/latest)
