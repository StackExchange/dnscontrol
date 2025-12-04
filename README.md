# DNSControl

[![StackExchange/dnscontrol/build](https://github.com/StackExchange/dnscontrol/actions/workflows/pr_build.yml/badge.svg)](https://github.com/StackExchange/dnscontrol/actions/workflows/pr_build.yml)
[![Google Group](https://img.shields.io/badge/google%20group-chat-green.svg)](https://groups.google.com/forum/#!forum/dnscontrol-discuss)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/StackExchange/dnscontrol)](https://pkg.go.dev/github.com/StackExchange/dnscontrol/v4)

[DNSControl](https://docs.dnscontrol.org/) is a system
for maintaining DNS zones.  It has two parts:
a domain specific language (DSL) for describing DNS zones plus
software that processes the DSL and pushes the resulting zones to
DNS providers such as Route53, Cloudflare, and Gandi.  It can send
the same DNS records to multiple providers.  It even generates
the most beautiful BIND zone files ever.  It runs anywhere Go runs (Linux, macOS,
Windows). The provider model is extensible, so more providers can be added.

Currently supported DNS providers:

- AdGuard Home
- Akamai Edge DNS
- Alibaba Cloud DNS (ALIDNS)
- AutoDNS
- AWS Route 53
- AXFR+DDNS
- Azure DNS
- Azure Private DNS
- BIND
- Bunny DNS
- CentralNic Reseller (CNR) - formerly RRPProxy
- Cloudflare
- ClouDNS
- CSC Global (*Experimental*)
- deSEC
- DigitalOcean
- DNS Made Easy
- DNSimple
- Domainnameshop (Domeneshop)
- Exoscale
- Fortigate
- Gandi
- Gcore
- Google DNS
- Hetzner
- HEXONET
- hosting.de
- Huawei Cloud DNS
- Hurricane Electric DNS
- INWX
- Joker
- Linode
- Loopia
- LuaDNS
- Microsoft Windows Server DNS Server
- Mythic Beasts
- Name.com
- Namecheap
- Netcup
- Netlify
- NS1
- Oracle Cloud
- OVH
- Packetframe
- Porkbun
- PowerDNS
- Realtime Register
- RWTH DNS-Admin
- Sakura Cloud
- SoftLayer
- TransIP
- Vercel
- Vultr

Currently supported Domain Registrars:

- AWS Route 53
- CentralNic Reseller (CNR) - formerly RRPProxy
- CSC Global
- DNSimple
- DNSOVERHTTPS
- Dynadot
- easyname
- Gandi
- HEXONET
- hosting.de
- Internet.bs
- INWX
- Loopia
- Name.com
- Namecheap
- OpenSRS
- OVH
- Porkbun
- Realtime Register

At Stack Overflow, we use this system to manage hundreds of domains
and subdomains across multiple registrars and DNS providers.

You can think of it as a DNS compiler.  The configuration files are
written in a DSL that looks a lot like JavaScript.  It is compiled
to an intermediate representation (IR).  Compiler back-ends use the
IR to update your DNS zones on services such as Route53, Cloudflare,
and Gandi, or systems such as BIND.

## An Example

`dnsconfig.js`:

```js
// define our registrar and providers
var REG_NAMECOM = NewRegistrar("name.com");
var r53 = NewDnsProvider("r53")

D("example.com", REG_NAMECOM, DnsProvider(r53),
  A("@", "1.2.3.4"),
  CNAME("www","@"),
  MX("@",5,"mail.myserver.com."),
  A("test", "5.6.7.8")
)
```

Running `dnscontrol preview` will talk to the providers (here name.com as registrar and route 53 as the dns host), and determine what changes need to be made.

Running `dnscontrol push` will make those changes with the provider and my dns records will be correctly updated.

The easiest way to run DNSControl is to use the Docker container:

```
docker run --rm -it -v "$(pwd):/dns"  ghcr.io/stackexchange/dnscontrol preview
```

See [Getting Started](https://docs.dnscontrol.org/getting-started/getting-started) page on documentation site to get started!

## Benefits

- **Less error-prone** than editing a BIND zone file.
- **More reproducible**  than clicking buttons on a web portal.
- **Easily switch between DNS providers:**  The DNSControl language is
  vendor-agnostic.  If you use it to maintain your DNS zone records,
  you can switch between DNS providers easily. In fact, DNSControl
  will upload your DNS records to multiple providers, which means you
  can test one while switching to another. We've switched providers 3
  times in three years and we've never lost a DNS record.
- **Apply CI/CD principles to DNS!**  At StackOverflow we maintain our
  DNSControl configurations in Git and use our CI system to roll out
  changes.  Keeping DNS information in a VCS means we have full
  history.  Using CI enables us to include unit-tests and
  system-tests.  Remember when you forgot to include a "." at the end
  of an MX record?  We haven't had that problem since we included a
  test to make sure Tom doesn't make that mistake... again.
- **Adopt (GitOps) PR-based updates.**  Allow developers to send updates as PRs,
  which you can review before you approve.
- **Variables save time!**  Assign an IP address to a constant and use the
  variable name throughout the file. Need to change the IP address
  globally? Just change the variable and "recompile."
- **Macros!**  Define your SPF records, MX records, or other repeated data
  once and re-use them for all domains.
- **Control Cloudflare from a single source of truth.**  Enable/disable
  Cloudflare proxying (the "orange cloud" button) directly from your
  DNSControl files.
- **Keep similar domains in sync** with transforms and other features.  If
  one domain is supposed to be a filtered version of another, this is
  easy to set up.
- **It is extendable!**  All the DNS providers are written as plugins.
  Writing new plugins is very easy.

## Installation

DNSControl can be installed via packages for macOS, Linux and Windows, or from source code. See the [official instructions](https://docs.dnscontrol.org/getting-started/getting-started#1-install-the-software).

## Via GitHub Actions (GHA)

See [dnscontrol-action](https://github.com/koenrh/dnscontrol-action) or [gacts/install-dnscontrol](https://github.com/gacts/install-dnscontrol).

## Deprecation warnings (updated 2025-11-21)

- **REV() will switch from RFC2317 to RFC4183 in v5.0.**  This is a breaking change. Warnings are output if your configuration is affected. No date has been announced for v5.0. See https://docs.dnscontrol.org/language-reference/top-level-functions/revcompat
- **NAMEDOTCOM, OPENSRS, and SOFTLAYER need maintainers!** These providers have no maintainer. Maintainers respond to PRs and fix bugs in a timely manner, and try to stay on top of protocol changes. Interested in being a hero and adopting them?  Contact tlimoncelli at stack overflow dot com.

## More info at our website

The website: [https://docs.dnscontrol.org/](https://docs.dnscontrol.org/)

The getting started guide: [https://docs.dnscontrol.org/getting-started/getting-started](https://docs.dnscontrol.org/getting-started/getting-started)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/StackExchange/dnscontrol.svg?variant=adaptive)](https://starchart.cc/StackExchange/dnscontrol)
