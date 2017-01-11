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
For this example we will use a domain registered with name.com, using their basic dns hosting.
The default name is `dnsconfig.js`:

{% highlight js %}
var registrar = NewRegistrar("name.com",NAMEDOTCOM);
var namecom = NewDnsProvider("name.com",NAMEDOTCOM);

D("example.com", registrar, DnsProvider(namecom),
  A("@", "1.2.3.4")
);
{%endhighlight%}

The second file is a json document to hold your api credentials. By default we use `providers.json`:

{% highlight json %}
{
  "name.com":{
    "apikey": "yourApiKeyFromName.com-klasjdkljasdlk235235235235",
    "apiuser": "yourUsername"
  }
}
{%endhighlight%}

You may modify these files to match your particular providers and domains. See [the javascript docs]({{site.github.url}}/js) for more details.

## 3. Run `dnscontrol preview`

This will print out a list of "corrections" that need to be performed. It will not actually make any changes.

## 4. Run `dnscontrol push`

This will actually perform the required changes with the various providers.