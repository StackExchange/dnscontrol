---
name: Internet.bs
title: Internet.bs Provider
layout: default
jsId: INTERNETBS
---
# Internet.bs Provider

Internet.bs has API to mange your DNS zones. But in current version of provider, it's not implemented. You can use Internet.bs only as registrar for now.


## Configuration
In your credentials file, you must provide your API key and account password 

{% highlight json %}
{
  "internetbs": {
    "api-key": "your-api-key",
    "password": "account-password"
  }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Internet.bs.

## Usage
Example Javascript:

{% highlight js %}
var REG_INERNETBS = NewRegistrar('internetbs', 'INTERNETBS');
var GCLOUD = NewDnsProvider("gcloud", "GCLOUD"); // Any provider

D("example.tld", REG_INERNETBS, DnsProvider(GCLOUD),
    A("test","1.2.3.4")
);
{% endhighlight %}

## Activation

Pay attention, you need to define white list of IP for API. But you always can change it on `My Profile > Reseller Settings`   
