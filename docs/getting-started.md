---
layout: default
---
# Getting Started

## 1. Get the binaries

You can either download the latest [github release](https://github.com/StackExchange/dnscontrol/releases), or build from the go source:

`go get github.com/StackExchange/dnscontrol`

## 2. Create files

The first file you will need is a javascript file to describe your domains.
Individual providers will vary slightly. See [the provider docs]({{site.github.url}}/provider-list) for specifics.
For this example we will use a single "BIND" provider that will generate zone files on our local file system.
The default name is `dnsconfig.js`:

{% highlight js %}
var registrar = NewRegistrar("none","NONE"); // no registrar
var bind = NewDnsProvider("bind","BIND");

D("example.com", registrar, DnsProvider(bind),
  A("@", "1.2.3.4")
);
{%endhighlight%}

You may modify this files to match your particular providers and domains. See [the javascript docs]({{site.github.url}}/js) and  [the provider docs]({{site.github.url}}/provider-list) for more details. If you are using other providers, you will likely need to make a `creds.json` file with api tokens and other account information.

## 3. Run `dnscontrol preview`

This will print out a list of "corrections" that need to be performed. It will not actually make any changes.

## 4. Run `dnscontrol push`

This will actually generate `zones/example.com.zone` for you. The bind provider is more configurable, and you can read more information [here.]({{site.github.url}}/providers/bind)
