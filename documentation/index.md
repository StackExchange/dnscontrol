{% hint style="info" %}
<span style="font-size: 21px; font-weight: 200;">DNSControl is an <a href="https://docs.dnscontrol.org/developer-info/opinions">opinionated</a> platform for seamlessly managing your DNS configuration across any number of DNS hosts, both in the cloud or in your own infrastructure. It manages all of the domains for the Stack Overflow network, and can do the same for you!</span>
{% endhint %}

# Try It

Want to jump right in? Follow our [quick start tutorial](getting-started.md) on a new domain or [migrate](migrating.md) an existing one. Read the [language spec](js.md) for more info.

# Use It

Take advantage of the advanced features. Use macros and variables for easier updates. Upload your zones to [multiple DNS providers](providers.md).

{% hint style="success" %}
* Maintain your DNS data as a high-level DS, with macros, and variables for easier updates.
* Super extensible! Plug-in architecture makes adding new DNS providers and Registrars easy!
* Eliminate vendor lock-in. Switch DNS providers easily, any time, with full fidelity.
* Reduce points of failure: Easily maintain dual DNS providers and easily drop one that is down.
* Supports 35+ [DNS Providers](providers.md) including [BIND](providers/bind.md), [AWS Route 53](providers/route53.md), [Google DNS](providers/gcloud.md), and [name.com](providers/namedotcom.md).
* [Apply CI/CD principles](ci-cd-gitlab.md) to DNS: Unit-tests, system-tests, automated deployment.
* All the benefits of Git (or any VCS) for your DNS zone data. View history. Accept PRs.
* Optimize DNS with [SPF optimizer](language-reference/domain/SPF_BUILDER.md). Detect too many lookups. Flatten includes.
* Runs on Linux, Windows, Mac, or any operating system supported by Go.
* Enable/disable Cloudflare proxying (the "orange cloud" button) directly from your DNSControl files.
* [Assign an IP address to a constant](examples.md#variables-for-common-ip-addresses) and use the variable name throughout the configuration. Need to change the IP address globally? Just change the variable and "recompile".
* Keep similar domains in sync with transforms, [macros](examples.md#macro-to-for-repeated-records), and variables.
{% endhint %}

# Get Involved

Join our [mailing list](https://groups.google.com/g/dnscontrol-discuss). We make it easy to contribute by using [GitHub](https://github.com/StackExchange/dnscontrol), you can make code changes with confidence thanks to extensive integration tests. The project is [newbie-friendly](https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html) so jump right in!
