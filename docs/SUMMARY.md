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

## Service Providers

* [Providers](provider-list.md)
* Service Providers
    * [Akamai Edge DNS](providers/akamaiedgedns.md)
    * [AutoDNS](providers/autodns.md)
    * [AXFR+DDNS](providers/axfrddns.md)
    * [Azure DNS](providers/azure_dns.md)
    * [BIND](providers/bind.md)
    * [Cloudflare](providers/cloudflareapi.md)
    * [ClouDNS](providers/cloudns.md)
    * [CSC Global](providers/cscglobal.md)
    * [deSEC](providers/desec.md)
    * [DigitalOcean](providers/digitalocean.md)
    * [DNSimple](providers/dnsimple.md)
    * [DNS Made Simple](providers/dnsmadeeasy.md)
    * [DNS-over-HTTPS](providers/dnsoverhttps.md)
    * [DOMAINNAMESHOP](providers/domainnameshop.md)
    * [easyname](providers/easyname.md)
    * [Gandi_v5](providers/gandi_v5.md)
    * [Google Cloud DNS](providers/gcloud.md)
    * [Hurricane Electric DNS](providers/hedns.md)
    * [Hetzner DNS Console](providers/hetzner.md)
    * [HEXONET](providers/hexonet.md)
    * [hosting.de](providers/hostingde.md)
    * [Internet.bs](providers/internetbs.md)
    * [INWX](providers/inwx.md)
    * [Linode](providers/linode.md)
    * [Microsoft DNS Server on Microsoft Windows Server](providers/msdns.md)
    * [Namecheap](providers/namecheap.md)
    * [Name.com](providers/namedotcom.md)
    * [Netcup](providers/netcup.md)
    * [NS1](providers/ns1.md)
    * [Oracle Cloud](providers/oracle.md)
    * [OVH](providers/ovh.md)
    * [Packetframe](providers/packetframe.md)
    * [PowerDNS](providers/powerdns.md)
    * [Amazon Route 53](providers/route53.md)
    * [RWTH DNS-Admin](providers/rwth.md)
    * [SoftLayer DNS](providers/softlayer.md)
    * [TransIP DNS](providers/transip.md)
    * [Vultr](providers/vultr.md)

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
* [Unit Testing DNS Data](unittests.md)
* [Useful code tricks](code-tricks.md)
* [Why CNAME/MX/NS targets require a "dot"](why-the-dot.md)

## Developer info

* [ALIAS Records](alias.md)
* [Bug Triage Process](bug-triage.md)
* [How to build and ship a release](release-engineering.md)
* [Bring-Your-Own-Secrets for automated testing](byo-secrets.md)
* [Writing new DNS providers](writing-providers.md)
* [Creating new DNS Resource Types (rtypes)](adding-new-rtypes.md)
* [TXT record testing](testing-txt-records.md)
* [Releases](https://github.com/StackExchange/dnscontrol/releases/latest)
