---
name: HEXONET
title: HEXONET Provider
layout: default
jsId: HEXONET
---
# HEXONET Provider

HEXONET is a leading developer and operator of domain names and DNS platforms.
Individual, service provider and registrars around the globe choose HEXONET for
domains and DNS because of our advanced technology, operational performance and
up-time, and most importantly for DNS expertise. DNSControl with HEXONET's DNS
marries DNS automation with an industry-leading DNS platform that supports DNSSEC,
PremiumDNS via Anycast Network, and nearly all of DNSControl's listed provider features.

This is based on API documents found at [https://wiki.hexonet.net/wiki/DNS_API](https://wiki.hexonet.net/wiki/DNS_API)

## Configuration

Please provide your HEXONET login data in your credentials file `creds.json` as follows:

{% highlight json %}
{
  "hexonet": {
    "apilogin": "your-hexonet-account-id",
    "apipassword": "your-hexonet-account-password",
    "apientity": "LIVE", // for the LIVE system; use "OTE" for the OT&E system
    "ipaddress": "172.31.3.16", // provide here your outgoing ip address
    "debugmode": "0", // set it to "1" to get debug output of the communication with our Backend System API
  }
}
{% endhighlight %}

Here a working example for our OT&E System:

{% highlight json %}
{
  "hexonet": {
    "apilogin": "test.user",
    "apipassword": "test.passw0rd",
    "apientity": "OTE",
    "debugmode": "0",
  }
}
{% endhighlight %}

NOTE: The above credentials are known to the public.

With the above hexonet entry in `creds.json`, you can run the
integration tests as follows:

    dnscontrol get-zones --format=nameonly hexonet HEXONET  all
    # Review the output.  Pick one domain and set HEXONET_DOMAIN.
    cd $GIT/dnscontrol/integrationTest
    export HEXONET_DOMAIN=a-b-c-movies.com       # Pick a domain name.
    export HEXONET_ENTITY=OTE
    export HEXONET_UID=test.user
    export HEXONET_PW=test.passw0rd
    go test -v -verbose -provider HEXONET

## Usage

Here's an example DNS Configuration `dnsconfig.js` using our provider module.
Even though it shows how you use us as Domain Registrar AND DNS Provider, we don't force you to do that.
You are free to decide if you want to use both of our provider technology or just one of them.

{% highlight javascript %}
// Providers:
var REG_HX = NewRegistrar('hexonet', 'HEXONET');
var DNS_HX = NewDnsProvider('hexonet', 'HEXONET');

// Set Default TTL for all RR to reflect our Backend API Default
// If you use additional DNS Providers, configure a default TTL
// per domain using the domain modifier DefaultTTL instead.
// also check this issue for [NAMESERVER TTL](https://github.com/StackExchange/dnscontrol/issues/176).
DEFAULTS(
    {"ns_ttl":"3600"},
    DefaultTTL(3600)
);

// Domains:
D('abhoster.com', REG_HX, DnsProvider(DNS_HX),
    NAMESERVER('ns1.ispapi.net'),
    NAMESERVER('ns2.ispapi.net'),
    NAMESERVER('ns3.ispapi.net'),
    NAMESERVER('ns4.ispapi.net'),
    A('elk1', '10.190.234.178'),
    A('test', '56.123.54.12')
);
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to HEXONET.

## get-zones

`dnscontrol get-zones` is implemented for this provider. The list
includes both basic and premier zones.

## New domains

If a dnszone does not exist in your HEXONET account, DNSControl will *not* automatically add it with the `dnscontrol push` or `dnscontrol preview` command. You'll need to do that via the control panel manually or using the command `dnscontrol create-domains`.
This is because it could lead to unwanted costs on customer-side that we want to avoid.

## Debug Mode

As shown in the configuration examples above, this can be activated on demand and it can be used to check the API commands send to our system.
In general this is thought for our purpose to have an easy way to dive into issues. But if you're interested what's going on, feel free to activate it.

## IP Filter

In case you have ip filter settings made for your HEXONET account, please provide your outgoing ip address as shown in the configuration examples above.
