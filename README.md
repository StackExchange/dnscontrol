# DNSControl

[![Build Status](https://travis-ci.org/StackExchange/dnscontrol.svg)](https://travis-ci.org/StackExchange/dnscontrol)

DNSControl is a system for maintaining DNS zones.  It has two parts:
a domain specific language (DSL) for describing DNS zones plus
software that processes the DSL and pushes the resulting zones to
DNS providers such as Route53, CloudFlare, and Gandi.  It can talk
to Microsoft ActiveDirectory and it generates the most beautiful
BIND zone files ever.  It run anywhere Go runs (Linux, macOS,
Windows).

At Stack Overflow, we use this system to manage hundreds of domains
and subdomains across multiple registrars and DNS providers.

You can think of it as a DNS compiler.  The configuration files are
written in a DSL that looks a lot like JavaScript.  It is compiled
to an intermediate representation (IR).  Compiler back-ends use the
IR to update your DNS zones on services such as Route53, CloudFlare,
and Gandi, or systems such as BIND and ActiveDirectory.

# Benefits

* Editing zone files is error-prone.  Clicking buttons on a web
page is irreproducible.
* Switching DNS providers becomes a no-brainer.  The DNSControl
language is vendor-agnostic.  If you use it to maintain your DNS
zone records, you can switch between DNS providers easily. In fact,
DNSControl will upload your DNS records to multiple providers, which
means you can test one while switching to another. We've switched
providers 3 times in three years and we've never lost a DNS record.
* Adopt CI/CD principles to DNS!  At StackOverflow we maintain our
DNSControl configurations in Git and use our CI system to roll out
changes.  Keeping DNS information in a VCS means we have full
history.  Using CI enables us to include unit-tests and system-tests.
Remember when you forgot to include a "." at the end of an MX record?
We haven't had that problem since we included a test to make sure
Tom doesn't make that mistake... again.
* Variables save time!  Assign an IP address to a constant and use
the variable name throughout the file. Need to change the IP address
globally? Just change the variable and "recompile."
* Macros!  Define your SPF records, MX records, or other repeated
data once and re-use them for all domains.
* Control CloudFlare from a single location.  Enable/disable
Cloudflare proxying (the "orange cloud" button) directly from your
DNSControl files.
* Keep similar domains in sync with transforms and other features.
If one domain is supposed to be the same
* It is extendable!  All the DNS providers are written as plugins.
Writing new plugins is very easy.

# Installation

`go get github.com/StackExchange/DNSControl`

or get prebuilt binaries from our github page.
