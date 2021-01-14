---
name: Oracle Cloud
title: Oracle Cloud Provider
layout: default
jsId: ORACLE
---
# Oracle Cloud Provider

## Configuration

Create an API key through the Oracle Cloud portal, and provide the user OCID, tenancy OCID, key fingerprint, region, and the contents of the private key.
The OCID of the compartment DNS resources should be put in can also optionally be provided.

{% highlight json %}
{
  "oracle": {
    "user_ocid": "$ORACLE_USER_OCID",
    "tenancy_ocid": "$ORACLE_TENANCY_OCID",
    "fingerprint": "$ORACLE_FINGERPRINT",
    "region": "$ORACLE_REGION",
    "private_key": "$ORACLE_PRIVATE_KEY",
    "compartment": "$ORACLE_COMPARTMENT"
  },
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Oracle Cloud.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE')
var ORACLE = NewDnsProvider("oracle", "ORACLE");

D("example.tld", REG_NONE, DnsProvider(ORACLE),
    NAMESERVER_TTL(86400),

    A("test","1.2.3.4")
);
{% endhighlight %}

