---
layout: default
---
# Getting Started

## 1. Get the binaries

You can either download the latest [github release](https://github.com/StackExchange/dnscontrol/releases), or build from the go source:

`go get github.com/StackExchange/dnscontrol`

If you are unfamiliar with Go, this will will download the source,
compile it, and install "dnscontrol" in your `bin` directory.

## 2. Create and test a sample configuration.

DNSControl has two primary configuration files.

1.  `dnsconfig.js` is the main configuration and defines providers,
DNS domains, and so on.

Start your `dnsconfig.js` file by downloading
[dnsconfig.js-example.txt]({{ site.github.url }}/assets/dnsconfig.js-example.txt))
and renaming it.

The file looks like:

{% highlight js %}

// Providers:

var REG_NONE = NewRegistrar('none', 'NONE');    // No registrar.
var DNS_BIND = NewDnsProvider('bind', 'BIND');  // ISC BIND.

// Domains:

D('example.com', REG_NONE, DnsProvider(DNS_BIND),
    A('@', '1.2.3.4')
);
{%endhighlight%}

There are 2 types of providers: a "Registrar" is who you register the domain with.
Normally `REG_NONE` is used, and DNSControl won't try to change anything at the registrar.
The `DnsProvider` is the service that actually provides DNS service (port 53) and
may be the same or different company. Even if both your Registrar and DnsProvider are
the same company, two different defintions must be included in `dnsconfig.js`.

2.  `creds.json` is a configuration file for storing credentials.
It is only needed if any providers require credentials (API keys,
usernames, passwords, etc.).

Start your `creds.json` file by downloading
[creds.json-example.txt]({{ site.github.url }}/assets/creds.json-example.txt))
and renaming it.

The file looks like:

{% highlight js %}
{
  "bind": {
  },
  "r53_ACCOUNTNAME": {
    "KeyId": "change_to_your_keyid",
    "SecretKey": "change_to_your_secretkey"
  }
}
{%endhighlight%}

Ignore the `r53_ACCOUNTNAME` section.  It is a placeholder and will be ignored. Later
if you use a provider that uses an API, you can use this as a template.

Note that `creds.json` is a JSON file, which is very strict about commas
and other formatting.  There are a few different ways to check for typos:

Python:

    python -m json.tool creds.json

jq:

    jq < creds.json

Lastly, create a subdirectory called `zones` in the same directory
as the configuration files.  (`mkdir zones`).  `zones` is where the
BIND provider writes the zonefiles it creates.

## 3. Test the sample files.

Before you edit the sample files, verify that the system is working.

First run `dnscontrol preview` and make sure that it completes with no errors.  The preview command
is the "dry run" mode that shows what changes need to be made and never makes any actual changes.  
It will use APIs if needed to find out what DNS entries currently exist.

It should look something like this:

{% highlight js %}

$ dnscontrol preview
Initialized 1 registrars and 1 dns service providers.
******************** Domain: example.com
----- Getting nameservers from: bind
----- DNS Provider: bind... 1 correction
#1: GENERATE_ZONEFILE: example.com
 (2 records)

----- Registrar: none
Done. 1 corrections.
{%endhighlight%}

Next run `dnscontrol push` to actually make the changes. In this
case, the change will be to create a zone file where one didn't
previously exist.

{% highlight js %}
$ dnscontrol push
Initialized 1 registrars and 1 dns service providers.
******************** Domain: example.com
----- Getting nameservers from: bind
----- DNS Provider: bind... 1 correction
#1: GENERATE_ZONEFILE: example.com
 (2 records)

CREATING ZONEFILE: zones/example.com.zone
SUCCESS!
----- Registrar: none
Done. 1 corrections.
{%endhighlight%}


## 4. Make a change.

Try making a change to `dnsconfig.js`. For example, change the IP
address of in `A('@', '1.2.3.4')` or add an additional A record.

In our case, we changed the IP address to 10.10.10.10. Previewing
our change looks like this:

{% highlight js %}
$ dnscontrol preview
Initialized 1 registrars and 1 dns service providers.
******************** Domain: example.com
----- Getting nameservers from: bind
----- DNS Provider: bind... 1 correction
#1: GENERATE_ZONEFILE: example.com
MODIFY A example.com: (1.2.3.4 300) -> (10.10.10.10 300)

----- Registrar: none
Done. 1 corrections.
{%endhighlight%}

Notice that it read the old zone file and was able to
produce a "diff" between the old `A` record and the new one.
If the zonefile didn't exist, the output would look different
because the zone file was being created from scratch.

Run `dnscontrol push` to see the system generate a new zone file.

Other providers use an API do do updates. In those cases
the individual changes will translate into API calls that
update the specific records.

Take a look at the `zones/example.com.zone` file.  It should
look like:

{% highlight js %}
$TTL 300
@                IN SOA   DEFAULT_NOT_SET. DEFAULT_NOT_SET. 1 3600 600 604800 1440
                 IN A     10.10.10.10
{%endhighlight%}

You can change the "DEFAULT_NOT_SET" text by following the documentation
for the [BIND provider]({{site.github.url}}/providers/bind) to set
the "master" and "mbox" settings.

## 5. Use your own domains

Now that we know the system is working for test data, try controlling
a real domain (or a test domain if you have one).

Set up the provider:  Add the providers's definition to `dnsconfig.js`
and list any credentials in `creds.json`.  Each provider is different.
See [the provider docs]({{site.github.url}}/provider-list) for
specifics.

Edit the domain: Add the `D()` entry for the domain, or repurpose
the `example.com` domain. Add the individual `A()`, `MX()` and other
records.

Run `dnscontrol preview` to test your work. It may take a few tries
to list all the DNS records that make up the domain.  When
preview shows no changes required, then you know you are at
feature parity.

The [Migrating]({{site.github.url}}/migrating) doc has advice
about converting from other systems.
You can manually create the `D()` statements, or you can
generate them automatically using the
[convertzone](https://github.com/StackExchange/dnscontrol/blob/master/misc/convertzone/README.md)
utility that is included in the DNSControl repo (it converts
BIND-style zone files to DNSControl's language).

Now you can make change to the domain(s)  and run `dnscontrol preview`
to check your work, then `dnscontrol push` to actually make your changes.
