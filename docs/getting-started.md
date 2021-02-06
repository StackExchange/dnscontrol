---
layout: default
title: Getting Started
---
# Getting Started


## 1. Install the software

## From source

DNSControl can be built with Go version 1.14 or higher.

The `go get` command will will download the source, compile it, and
install `dnscontrol` in your `$GOBIN` directory.

To install, simply run

    GO111MODULE=on go get github.com/StackExchange/dnscontrol/v3

To download the source

    git clone github.com/StackExchange/dnscontrol

If these don't work, more info is in [#805](https://github.com/StackExchange/dnscontrol/issues/805).

---


## Via packages

Get prebuilt binaries from [github releases](https://github.com/StackExchange/dnscontrol/releases/latest)

Alternatively, on Mac you can install it using homebrew:

`brew install dnscontrol`

## Via [docker](https://hub.docker.com/r/stackexchange/dnscontrol/)

```
docker run --rm -it -v $(pwd)/dnsconfig.js:/dns/dnsconfig.js -v $(pwd)/creds.json:/dns/creds.json stackexchange/dnscontrol dnscontrol preview
```


## 2. Create a place for the config files.

Create a directory where you'll be storing your configuration files.
We highly recommend storing these files in a Git repo, but for
simple tests anything will do.

Note: Do **not** store your creds.json file in Git unencrypted.
That is unsafe. Add `creds.json` to your
`.gitignore` file as a precaution.  This file should be encrypted
using something
like [git-crypt](https://www.agwa.name/projects/git-crypt) or
[Blackbox](https://github.com/StackExchange/blackbox).

Create a subdirectory called `zones` in the same directory as the
configuration files.  (`mkdir zones`).  `zones` is where the BIND
provider writes the zonefiles it creates. Even if you don't
use BIND, it is useful for testing.


## 3. Create the initial `dnsconfig.js`

`dnsconfig.js` is the main configuration and defines providers, DNS
domains, and so on.

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

You may modify this file to match your particular providers and domains. See [the javascript docs]({{site.github.url}}/js) and  [the provider docs]({{site.github.url}}/provider-list) for more details.
If you are using other providers, you will likely need to make a `creds.json` file with api tokens and other account information. For example, to use both name.com and Cloudflare, you would have:

{% highlight js %}
{
  "cloudflare":{ // provider name to be used in dnsconfig.js
    "apitoken": "token" // API token
  },
  "namecom":{ // provider name to be used in dnsconfig.js
    "apikey": "key", // API Key
    "apiuser": "username" // username for name.com
  }
}
{%endhighlight%}

There are 2 types of providers:

A "Registrar" is who you register the domain with.  Start with
`REG_NONE`, which is a provider that never talks to or updates the
registrar.  You can define your registrar later when you want to
use advanced features.

The `DnsProvider` is the service that actually provides DNS service
(port 53) and may be the same or different company. Even if both
your Registrar and DnsProvider are the same company, two different
definitions must be included in `dnsconfig.js`.


## 4. Create the initial `creds.json`

`creds.json` stores credentials and a few global settings.
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

Ignore the `r53_ACCOUNTNAME` section.  It is a placeholder and will be ignored. You
can use it later when you define your first set of API credentials.

Note that `creds.json` is a JSON file. JSON is very strict about commas
and other formatting.  There are a few different ways to check for typos:

Python:

    python -m json.tool creds.json

jq:

    jq < creds.json

FYI: `creds.json` fields can be read from an environment variable. The field must begin with a `$` followed by the variable name. No other text. For example:

    "apikey": "$GANDI_V5_APIKEY",

## 5. Test the sample files.

Before you edit the sample files, verify that the system is working.

First run `dnscontrol preview` and make sure that it completes with
no errors.  The preview command is the "dry run" mode that shows
what changes need to be made and never makes any actual changes.
It will use APIs if needed to find out what DNS entries currently
exist.

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


## 6. Make a change.

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

Notice that it read the old zone file and was able to produce a
"diff" between the old `A` record and the new one.  If the zonefile
didn't exist, the output would look different because the zone file
was being created from scratch.

Run `dnscontrol push` to see the system generate a new zone file.

Other providers use an API do do updates. In those cases the
individual changes will translate into API calls that update the
specific records.

Take a look at the `zones/example.com.zone` file.  It should look
like:

{% highlight js %}
$TTL 300
@                IN SOA   DEFAULT_NOT_SET. DEFAULT_NOT_SET. 1 3600 600 604800 1440
                 IN A     10.10.10.10
{%endhighlight%}

You can change the "DEFAULT_NOT_SET" text by following the documentation
for the [BIND provider]({{site.github.url}}/providers/bind) to set
the "master" and "mbox" settings.  Try that now.


## 7. Use your own domains

Now that we know the system is working for test data, try controlling
a real domain (or a test domain if you have one).

Set up the provider:  Add the providers's definition to `dnsconfig.js`
and list any credentials in `creds.json`.  Each provider is different.
See [the provider docs]({{site.github.url}}/provider-list) for
specifics.

Edit the domain: Add the `D()` entry for the domain, or repurpose
the `example.com` domain.  Add individual `A()`, `MX()` and other
records as needed.  Remember that the first parameter to `D()` is
always a Registrar.

Run `dnscontrol preview` to test your work. It may take a few tries
to list all the DNS records that make up the domain.  When preview
shows no changes required, then you know you are at feature parity.

The [Migrating]({{site.github.url}}/migrating) doc has advice
about converting from other systems.
You can manually create the `D()` statements, or you can
generate them automatically using the
[dnscontrol get-zones]({{site.github.url}}/get-zones)
command to import the zone from (most) providers and output it as code
that can be added to `dnsconfig.js` and used with very little
modification.

Now you can make change to the domain(s)  and run `dnscontrol preview`


## 8. Production Advice

If you are going to use this in production, we highly recommend the following:

* Store the configuration files in Git.
* Encrypt the `creds.json` file before storing it in Git. Do NOT store
  API keys or other credentials without encrypting them.
* Use a CI/CD tool like Jenkins/CircleCI/Github Actions/etc. to automatically push DNS changes.
* Join the DNSControl community. File [issues and PRs](https://github.com/StackExchange/dnscontrol).
