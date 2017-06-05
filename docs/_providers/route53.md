---
name: Route 53
layout: default
jsId: ROUTE53
---
# Amazon Route 53 Provider

## Configuration

By default, you can configure aws setting like the [go sdk configuration](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html). For example you can use environment variables:
```
$ export AWS_ACCESS_KEY_ID=XXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=YYYYYYYYY
```

It is also possible to specify an aws access key in the providers config json file:

{% highlight json %}
{
 "r53":{
      "KeyId": "your-aws-key",
      "SecretKey": "your-aws-secret-key"
 }
}
{% endhighlight %}

## Metadata

This provider does not recognize any special metadata fields unique to route 53.

## Usage

Example javascript:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_NAMECOM, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}

## Activation

DNSControl depends on a standard [aws access key](https://aws.amazon.com/developers/access-keys/) with permission to list, create and update hosted zones.

## New domains

If a domain does not exist in your Route53 account, DNSControl
will *not* automatically add it. You can do that either manually
via the control panel, or via the command `dnscontrol create-domains`
command.
