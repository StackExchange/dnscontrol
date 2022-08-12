# DNSControl

[![CircleCI](https://circleci.com/gh/StackExchange/dnscontrol/tree/master.svg?style=svg)](https://circleci.com/gh/StackExchange/dnscontrol/tree/master)
[![Gitter chat](https://badges.gitter.im/dnscontrol/Lobby.png)](https://gitter.im/dnscontrol/Lobby)
[![Google Group](https://img.shields.io/badge/google%20group-chat-green.svg)](https://groups.google.com/forum/#!forum/dnscontrol-discuss)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/StackExchange/dnscontrol)](https://pkg.go.dev/github.com/StackExchange/dnscontrol)
[![CircleCI](https://dl.circleci.com/insights-snapshot/gh/StackExchange/dnscontrol/master/build/badge.svg?window=30d)](https://app.circleci.com/insights/github/StackExchange/dnscontrol/workflows/build/overview?branch=master&reporting-window=last-30-days&insights-snapshot=true)

[DNSControl](https://stackexchange.github.io/dnscontrol/) is a system
for maintaining DNS zones.  It has two parts:
a domain specific language (DSL) for describing DNS zones plus
software that processes the DSL and pushes the resulting zones to
DNS providers such as Route53, Cloudflare, and Gandi.  It can send
the same DNS records to multiple providers.  It even generates
the most beautiful BIND zone files ever.  It runs anywhere Go runs (Linux, macOS,
Windows). The provider model is extensible, so more providers can be added.

Currently supported DNS providers:

- AWS Route 53
- AXFR+DDNS
- Active Directory (Deprecated, see Microsoft DNS)
- Akamai Edge DNS
- AutoDNS
- Azure DNS
- BIND
- ClouDNS
- Cloudflare
- deSEC
- DNS Made Easy
- DNSimple
- DigitalOcean
- DomainNameShop (domeneshop)
- Exoscale
- Gandi
- Google DNS
- Hetzner
- HEXONET
- hosting.de
- Hurricane Electric DNS
- INWX
- Linode
- Microsoft Windows Server DNS Server
- NS1
- Name.com
- Namecheap
- Netcup
- OVH
- OctoDNS
- Oracle Cloud
- Packetframe
- PowerDNS
- RWTH DNS-Admin
- SoftLayer
- TransIP
- Vultr

Currently supported Domain Registrars:

- AWS Route 53
- CSC Global
- DNSOVERHTTPS
- easyname
- Gandi
- HEXONET
- hosting.de
- Internet.bs
- INWX
- Name.com
- Namecheap
- OVH
- OpenSRS

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
var namecom = NewRegistrar("name.com", "NAMEDOTCOM");
var r53 = NewDnsProvider("r53", "ROUTE53")

D("example.com", namecom, DnsProvider(r53),
  A("@", "1.2.3.4"),
  CNAME("www","@"),
  MX("@",5,"mail.myserver.com."),
  A("test", "5.6.7.8")
)
```

Running `dnscontrol preview` will talk to the providers (here name.com as registrar and route 53 as the dns host), and determine what changes need to be made.

Running `dnscontrol push` will make those changes with the provider and my dns records will be correctly updated.

See [Getting Started](https://stackexchange.github.io/dnscontrol/getting-started) page on documentation site.

## Benefits

- **Less error-prone** than editing a BIND zone file.
- **More reproducible**  than clicking buttons on a web portal.
- **Easily switch between DNS providers:**  The DNSControl language is
  vendor-agnostic.  If you use it to maintain your DNS zone records,
  you can switch between DNS providers easily. In fact, DNSControl
  will upload your DNS records to multiple providers, which means you
  can test one while switching to another. We've switched providers 3
  times in three years and we've never lost a DNS record.
- **Adopt CI/CD principles to DNS!**  At StackOverflow we maintain our
  DNSControl configurations in Git and use our CI system to roll out
  changes.  Keeping DNS information in a VCS means we have full
  history.  Using CI enables us to include unit-tests and
  system-tests.  Remember when you forgot to include a "." at the end
  of an MX record?  We haven't had that problem since we included a
  test to make sure Tom doesn't make that mistake... again.
- **Adopt PR-based updates.**  Allow developers to send updates as PRs,
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

### From source

DNSControl can be built with Go version 1.16 or higher.

The `go get` command will download the source, compile it, and
install `dnscontrol` in your `$GOBIN` directory.

To install, simply run

```shell
GO111MODULE=on go install github.com/StackExchange/dnscontrol/v3@latest
```

To download the source

```shell
git clone https://github.com/StackExchange/dnscontrol.git
```

If these don't work, more info is in [#805](https://github.com/StackExchange/dnscontrol/issues/805).

---

### Via packages

Get prebuilt binaries from [github releases](https://github.com/StackExchange/dnscontrol/releases/latest)

Alternatively, on Mac you can install it using homebrew:

`brew install dnscontrol`

## Via [docker](https://hub.docker.com/r/stackexchange/dnscontrol/)

```bash
docker run --rm -it -v $(pwd)/dnsconfig.js:/dns/dnsconfig.js -v $(pwd)/creds.json:/dns/creds.json stackexchange/dnscontrol dnscontrol preview
```

The documentation can be viewed via Docker:

```bash
cd docs
docker run --rm -it --volume="$PWD:/srv/jekyll" --volume="$PWD/vendor/bundle:/usr/local/bundle" --env JEKYLL_ENV=production jekyll/jekyll:3.8 jekyll build -V
# Open docs/_site/index.html in your web browser to see the results.
# (Note: The preview isn't perfect. Links that use the site.github.url variable won't work.
```

## Via Github Actions (GHA)

See [dnscontrol-action](https://github.com/koenrh/dnscontrol-action)

## Depreciation warnings (updated 2022-06-04)

- **Call for new volunteer maintainers for NAMEDOTCOM, and SOFTLAYER.** These providers have no maintainer. Maintainers respond to PRs and fix bugs in a timely manner, and try to stay on top of protocol changes.
- **ACME/Let's Encrypt support is frozen and will be removed after December 31, 2022.**  The `get-certs` command (renews certs via Let's Encrypt) has no maintainer. There are other projects that do a better job. If you don't use this feature, please do not start. If you do use this feature, please plan on migrating to something else.  See discussion in [issues/1400](https://github.com/StackExchange/dnscontrol/issues/1400)
- **Provider OCTODNS is frozen and will be removed after Nov 1, 2022.** It was written as a joke and nobody laughed. It's time to remove the code and move on.
- **convertzone is frozen and will be removed after June 30, 2022.** The `convertzone` stand-alone program is replaced by the `get-zone` subcommand of DNSControl. It does everything `convertzone` did and more.
- **Provider ACTIVEDIRECTORY_PS is frozen and will be removed after June 30, 2022.** It is replaced by MSDNS which is 100% feature compatible and works better. (We believe nobody uses it.)
- **get-zones syntax changes in v3.16** Starting in v3.16, the command line arguments for `dnscontrol get-zones` changes. For backwards compatibility change `provider` to `-`. See documentation for details.

## More info at our web site

The website: [https://stackexchange.github.io/dnscontrol/](https://stackexchange.github.io/dnscontrol/)

The getting started guide: [https://stackexchange.github.io/dnscontrol/getting-started](https://stackexchange.github.io/dnscontrol/getting-started)
