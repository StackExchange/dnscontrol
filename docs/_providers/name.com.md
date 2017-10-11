---
name: Name.com
title: Name.com Provider
layout: default
jsId: NAMEDOTCOM
---

# Name.com Provider

## Configuration
In your credentials file you must provide your name.com api username and access token:

{% highlight json %}
{
  "name.com":{
    "apikey": "yourApiKeyFromName.com",
    "apiuser": "yourUsername"
  }
}
{% endhighlight %}

There is another key name `apiurl` but it is optional and defaults to the correct value. If you want to use the test environment ("OT&E"), then add this:

    "apiurl": "https://api.dev.name.com",

## Metadata
This provider does not recognize any special metadata fields unique to name.com.

## Usage
**Example Javascript (DNS hosted with name.com):**

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var NAMECOM = NewDnsProvider("name.com","NAMEDOTCOM");

D("example.tld", REG_NAMECOM, DnsProvider(NAMECOM),
    A("test","1.2.3.4")
);
{%endhighlight%}


**Example Javascript (Registrar only. DNS hosted elsewhere):**

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_NAMECOM, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}

{% include alert.html text="Note: name.com does not allow control over the NS records of your zones via the api. It is not recommended to use name.com's dns provider unless it is your only dns host." %}

## Activation
In order to activate API functionality on your Name.com account, you must apply to the API program. The application form is [located here](https://www.name.com/reseller/apply). It usually takes a few days to get a response. After you are accepted, you should receive your API token via email.