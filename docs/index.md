---
title: DNSControl
---

{% hint style="info" %}
<span style="font-size: 21px; font-weight: 200;">DNSControl is an [opinionated](opinions.md) platform for seamlessly managing your DNS configuration across any number of DNS hosts, both in the cloud or in your own infrastructure. It manages all of the domains for the Stack Overflow network, and can do the same for you!</span>
{% endhint %}

# Try It

Want to jump right in? Follow our [quick start tutorial](getting-started.md) on a new domain or [migrate](migrating.md) an existing one. Read the [language spec](js.md) for more info.

# Use It

Take advantage of the advanced features. Use macros and variables for easier updates. Upload your zones to [multiple DNS providers](provider-list.md).

{% hint style="success" %}
* Maintain your DNS data as a high-level DS, with macros, and variables for easier updates.
* Super extensible! Plug-in architecture makes adding new DNS providers and Registrars easy!
* Eliminate vendor lock-in. Switch DNS providers easily, any time, with full fidelity.
* Reduce points of failure: Easily maintain dual DNS providers and easily drop one that is down.
* Supports 35+ [DNS Providers](provider-list.md) including [BIND](_providers/bind.md), [AWS Route 53](_providers/route53.md), [Google DNS](_providers/gcloud.md), and [name.com](_providers/namedotcom.md).
* [Apply CI/CD principles](ci-cd-gitlab.md) to DNS: Unit-tests, system-tests, automated deployment.
* All the benefits of Git (or any VCS) for your DNS zone data. View history. Accept PRs.
* Optimize DNS with [SPF optimizer](_functions/record/SPF_BUILDER.md). Detect too many lookups. Flatten includes.
* Runs on Linux, Windows, Mac, or any operating system supported by Go.
* Enable/disable Cloudflare proxying (the "orange cloud" button) directly from your DNSControl files.
* [Assign an IP address to a constant](examples#variables-for-common-ip-addresses) and use the variable name throughout the configuration. Need to change the IP address globally? Just change the variable and "recompile".
* Keep similar domains in sync with transforms, [macros](examples#macro-to-for-repeated-records), and variables.
{% endhint %}

# Get Involved

Join our [mailing list](https://groups.google.com/g/dnscontrol-discuss). We make it easy to contribute by using [GitHub](https://github.com/StackExchange/dnscontrol), you can make code changes with confidence thanks to extensive integration tests. The project is [newbie-friendly](https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html) so jump right in!

## Getting Started

Information for new users and the curious.

* [Getting Started](getting-started.md): A walk-through of the basics
* [Providers](provider-list.md): Which DNS providers are supported
* [Examples](examples.md): The DNSControl language by example
* [Migrating](migrating.md): Migrating zones to DNSControl

## Commands

DNSControl sub-commands and options.

* [creds.json](creds-json.md): creds.json file format
* [check-creds](check-creds.md): Verify credentials
* [get-zones](get-zones.md): Query a provider for zone info
* [get-certs](get-certs.md): Renew SSL/TLS certs (DEPRECATED)

## Reference

Language resources and procedures.

* [Language Reference](js.md): Description of the DNSControl language (DSL)
* [Aliases](alias.md): ALIAS/ANAME records
* [SPF Optimizer](_functions/record/SPF_BUILDER.md): Optimize your SPF records
* [CAA Builder](_functions/record/CAA_BUILDER.md): Build CAA records the easy way

## Advanced features

Take advantage of DNSControl's unique features.

* [Why CNAME/MX/NS targets require a trailing "dot"](why-the-dot.md)
* [Testing](unittests.md): Unit Testing for you DNS Data
* [Notifications](notifications.md): Web-hook for changes
* [Code Tricks](code-tricks.md): Safely use macros and loops.
* [CLI variables](cli-variables.md): Passing variables from CLI to JS
* [Nameservers & delegation](nameservers.md): Many examples.
* [Gitlab CI/CD example](ci-cd-gitlab.md).

## Developer Info

It is easy to add features and new providers to DNSControl. The code is very modular and easy to modify. There are extensive integration tests that make it easy to boldly make changes with confidence that you'll know if anything is broken. Our mailing list is friendly. Afraid to make your first PR? We'll gladly mentor you through the process. Many major code contributions have come from [first-time Go users](https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html)!

* GitHub [StackExchange/dnscontrol](https://github.com/StackExchange/dnscontrol): Get the source!
* Mailing list: [dnscontrol-discuss](https://groups.google.com/g/dnscontrol-discuss): The friendly best place to ask questions and propose new features
* [Bug Triage](bug-triage.md): How bugs are triaged
* [Release Engineering](release-engineering.md): How to build and ship a release
* [Bring-Your-Own-Secrets](byo-secrets.md): Automate tests
* [Step-by-Step Guide](writing-providers.md): Writing Providers: How to write a DNS or Registrar Provider
* [Step-by-Step Guide](adding-new-rtypes.md): Adding new DNS rtypes: How to add a new DNS record type

Icons made by Freepik from [www.flaticon.com](https://www.flaticon.com)
