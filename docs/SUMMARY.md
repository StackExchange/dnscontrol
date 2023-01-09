# Table of contents

* [Introduction to DNSControl](index.md)

## Getting Started

* [Overview](getting-started.md)
* [Examples](examples.md)
* [Migrating zones to DNSControl](migrating.md)

## Language Reference

* [JavaScript DSL](js.md)
* Domain Modifiers
    * [A](_functions/domain/A.md)
    * [AAAA](_functions/domain/AAAA.md)
    * [ALIAS](_functions/domain/ALIAS.md)
    * [AUTODNSSEC_OFF](_functions/domain/AUTODNSSEC_OFF.md)
    * [AUTODNSSEC_ON](_functions/domain/AUTODNSSEC_ON.md)
    * [CAA](_functions/domain/CAA.md)
    * [CNAME](_functions/domain/CNAME.md)
    * [DS](_functions/domain/DS.md)
    * [DefaultTTL](_functions/domain/DefaultTTL.md)
    * [DnsProvider](_functions/domain/DnsProvider.md)
    * [FRAME](_functions/domain/FRAME.md)
    * [IGNORE](_functions/domain/IGNORE.md)
    * [IGNORE_NAME](_functions/domain/IGNORE_NAME.md)
    * [IGNORE_TARGET](_functions/domain/IGNORE_TARGET.md)
    * [IMPORT_TRANSFORM](_functions/domain/IMPORT_TRANSFORM.md)
    * [INCLUDE](_functions/domain/INCLUDE.md)
    * [MX](_functions/domain/MX.md)
    * [NAMESERVER](_functions/domain/NAMESERVER.md)
    * [NAMESERVER_TTL](_functions/domain/NAMESERVER_TTL.md)
    * [NO_PURGE](_functions/domain/NO_PURGE.md)
    * [NS](_functions/domain/NS.md)
    * [PTR](_functions/domain/PTR.md)
    * [PURGE](_functions/domain/PURGE.md)
    * [SOA](_functions/domain/SOA.md)
    * [SRV](_functions/domain/SRV.md)
    * [SSHFP](_functions/domain/SSHFP.md)
    * [TLSA](_functions/domain/TLSA.md)
    * [TXT](_functions/domain/TXT.md)
    * [URL](_functions/domain/URL.md)
    * [URL301](_functions/domain/URL301.md)
    * Service Provider specific
        * Akamai Edge Dns
            * [AKAMAICDN](_functions/domain/AKAMAICDN.md)
        * Azure DNS
            * [AZURE_ALIAS](_functions/domain/AZURE_ALIAS.md)
        * Cloudflare DNS
            * [CF_REDIRECT](_functions/domain/CF_REDIRECT.md)
            * [CF_TEMP_REDIRECT](_functions/domain/CF_TEMP_REDIRECT.md)
            * [CF_WORKER_ROUTE](_functions/domain/CF_WORKER_ROUTE.md)
        * Amazon Route 53
            * [R53_ALIAS](_functions/domain/R53_ALIAS.md)
* Record Modifiers
    * [CAA_BUILDER](_functions/record/CAA_BUILDER.md)
    * [DMARC_BUILDER](_functions/record/DMARC_BUILDER.md)
    * [SPF_BUILDER](_functions/record/SPF_BUILDER.md)
    * [TTL](_functions/record/TTL.md)
    * Service Provider specific
        * Amazon Route 53
            * [R53_ZONE](_functions/record/R53_ZONE.md)
* Top Level Functions
    * [D](_functions/global/D.md)
    * [DEFAULTS](_functions/global/DEFAULTS.md)
    * [DOMAIN_ELSEWHERE](_functions/global/DOMAIN_ELSEWHERE.md)
    * [DOMAIN_ELSEWHERE_AUTO](_functions/global/DOMAIN_ELSEWHERE_AUTO.md)
    * [D_EXTEND](_functions/global/D_EXTEND.md)
    * [FETCH](_functions/global/FETCH.md)
    * [IP](_functions/global/IP.md)
    * [NewDnsProvider](_functions/global/NewDnsProvider.md)
    * [NewRegistrar](_functions/global/NewRegistrar.md)
    * [PANIC](_functions/global/PANIC.md)
    * [REV](_functions/global/REV.md)
    * [getConfiguredDomains](_functions/global/getConfiguredDomains.md)
    * [require](_functions/global/require.md)
    * [require_glob](_functions/global/require_glob.md)
* [Why CNAME/MX/NS targets require a "dot"](why-the-dot.md)

## Service Providers

* [Providers](provider-list.md)
* Service Providers
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
    * [NS1](_providers/ns1.md)
    * [Oracle Cloud](_providers/oracle.md)
    * [OVH](_providers/ovh.md)
    * [Packetframe](_providers/packetframe.md)
    * [PowerDNS](_providers/powerdns.md)
    * [Amazon Route 53](_providers/route53.md)
    * [RWTH DNS-Admin](_providers/rwth.md)
    * [SoftLayer DNS](_providers/softlayer.md)
    * [TransIP DNS](_providers/transip.md)
    * [Vultr](_providers/vultr.md)

## Commands

* [creds.json](creds-json.md)
* [Check-Creds subcommand](check-creds.md)
* [Get-Zones subcommand](get-zones.md)
* [Let's Encrypt Certificate generation](get-certs.md)

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
