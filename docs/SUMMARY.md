# Table of contents

* [Introduction to DNSControl](index.md)

## Getting Started

* [Overview](getting-started.md)
* [Examples](examples.md)
* [Migrating zones to DNSControl](migrating.md)

## Language Reference

* [JavaScript DSL](js.md)
* Domain Modifiers
    * [A](\_functions/domain/A.md)
    * [AAAA](\_functions/domain/AAAA.md)
    * [ALIAS](\_functions/domain/ALIAS.md)
    * [AUTODNSSEC\_OFF](\_functions/domain/AUTODNSSEC\_OFF.md)
    * [AUTODNSSEC\_ON](\_functions/domain/AUTODNSSEC\_ON.md)
    * [CAA](\_functions/domain/CAA.md)
    * [CNAME](\_functions/domain/CNAME.md)
    * [DS](\_functions/domain/DS.md)
    * [DefaultTTL](\_functions/domain/DefaultTTL.md)
    * [DnsProvider](\_functions/domain/DnsProvider.md)
    * [FRAME](\_functions/domain/FRAME.md)
    * [IGNORE](\_functions/domain/IGNORE.md)
    * [IGNORE\_NAME](\_functions/domain/IGNORE\_NAME.md)
    * [IGNORE\_TARGET](\_functions/domain/IGNORE\_TARGET.md)
    * [IMPORT\_TRANSFORM](\_functions/domain/IMPORT\_TRANSFORM.md)
    * [INCLUDE](\_functions/domain/INCLUDE.md)
    * [MX](\_functions/domain/MX.md)
    * [NAMESERVER](\_functions/domain/NAMESERVER.md)
    * [NAMESERVER\_TTL](\_functions/domain/NAMESERVER\_TTL.md)
    * [NO\_PURGE](\_functions/domain/NO\_PURGE.md)
    * [NS](\_functions/domain/NS.md)
    * [PTR](\_functions/domain/PTR.md)
    * [PURGE](\_functions/domain/PURGE.md)
    * [SOA](\_functions/domain/SOA.md)
    * [SRV](\_functions/domain/SRV.md)
    * [SSHFP](\_functions/domain/SSHFP.md)
    * [TLSA](\_functions/domain/TLSA.md)
    * [TXT](\_functions/domain/TXT.md)
    * [URL](\_functions/domain/URL.md)
    * [URL301](\_functions/domain/URL301.md)
    * Service Provider specific
        * Akamai Edge Dns
            * [AKAMAICDN](\_functions/domain/AKAMAICDN.md)
        * Azure DNS
            * [AZURE\_ALIAS](\_functions/domain/AZURE\_ALIAS.md)
        * Cloudflare DNS
            * [CF\_REDIRECT](\_functions/domain/CF\_REDIRECT.md)
            * [CF\_TEMP\_REDIRECT](\_functions/domain/CF\_TEMP\_REDIRECT.md)
            * [CF\_WORKER\_ROUTE](\_functions/domain/CF\_WORKER\_ROUTE.md)
        * Amazon Route 53
            * [R53\_ALIAS](\_functions/domain/R53\_ALIAS.md)
* Record Modifiers
    * [CAA\_BUILDER](\_functions/record/CAA\_BUILDER.md)
    * [DMARC\_BUILDER](\_functions/record/DMARC\_BUILDER.md)
    * [SPF\_BUILDER](\_functions/record/SPF\_BUILDER.md)
    * [TTL](\_functions/record/TTL.md)
    * Service Provider specific
        * Amazon Route 53
            * [R53\_ZONE](\_functions/record/R53\_ZONE.md)
* Top Level Functions
    * [D](\_functions/global/D.md)
    * [DEFAULTS](\_functions/global/DEFAULTS.md)
    * [DOMAIN\_ELSEWHERE](\_functions/global/DOMAIN\_ELSEWHERE.md)
    * [DOMAIN\_ELSEWHERE\_AUTO](\_functions/global/DOMAIN\_ELSEWHERE\_AUTO.md)
    * [D\_EXTEND](\_functions/global/D\_EXTEND.md)
    * [FETCH](\_functions/global/FETCH.md)
    * [IP](\_functions/global/IP.md)
    * [NewDnsProvider](\_functions/global/NewDnsProvider.md)
    * [NewRegistrar](\_functions/global/NewRegistrar.md)
    * [PANIC](\_functions/global/PANIC.md)
    * [REV](\_functions/global/REV.md)
    * [getConfiguredDomains](\_functions/global/getConfiguredDomains.md)
    * [require](\_functions/global/require.md)
    * [require\_glob](\_functions/global/require\_glob.md)

## Service Providers

* [Providers](provider-list.md)
* [Service Providers](providers/README.md)
    * [Akamai Edge DNS Provider](providers/akamaiedgedns.md)
    * [AutoDNS Provider](providers/autodns.md)
    * [AXFR+DDNS Provider](providers/axfrddns.md)
    * [Azure DNS Provider](providers/azure\_dns.md)
    * [BIND Provider](providers/bind.md)
    * [Cloudflare Provider](providers/cloudflareapi.md)
    * [ClouDNS Provider](providers/cloudns.md)
    * [CSC Global Provider](providers/cscglobal.md)
    * [deSEC Provider](providers/desec.md)
    * [DigitalOcean Provider](providers/digitalocean.md)
    * [DNSimple Provider](providers/dnsimple.md)
    * [DNS Made Simple Provider](providers/dnsmadeeasy.md)
    * [DNS-over-HTTPS Provider](providers/dnsoverhttps.md)
    * [DOMAINNAMESHOP Provider](providers/domainnameshop.md)
    * [easyname Provider](providers/easyname.md)
    * [Gandi\_v5 Provider](providers/gandi\_v5.md)
    * [Google Cloud DNS Provider](providers/gcloud.md)
    * [Hurricane Electric DNS Provider](providers/hedns.md)
    * [Hetzner DNS Console Provider](providers/hetzner.md)
    * [HEXONET Provider](providers/hexonet.md)
    * [hosting.de Provider](providers/hostingde.md)
    * [Internet.bs Provider](providers/internetbs.md)
    * [INWX](providers/inwx.md)
    * [Linode Provider](providers/linode.md)
    * [Microsoft DNS Server on Microsoft Windows Server](providers/msdns.md)
    * [Namecheap Provider](providers/namecheap.md)
    * [Name.com Provider](providers/namedotcom.md)
    * [Netcup Provider](providers/netcup.md)
    * [NS1 Provider](providers/ns1.md)
    * [Oracle Cloud Provider](providers/oracle.md)
    * [OVH Provider](providers/ovh.md)
    * [Packetframe Provider](providers/packetframe.md)
    * [PowerDNS Provider](providers/powerdns.md)
    * [Amazon Route 53 Provider](providers/route53.md)
    * [RWTH DNS-Admin Provider](providers/rwth.md)
    * [SoftLayer DNS Provider](providers/softlayer.md)
    * [TransIP DNS Provider](providers/transip.md)
    * [Vultr Provider](providers/vultr.md)

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
