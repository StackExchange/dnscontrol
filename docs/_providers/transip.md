---
name: TransIP DNS
title: TransIP DNS Provider
layout: default
jsId: TRANSIP
---

# TransIP DNS Provider

## Configuration

In your providers config json file you must include your TransIP credentials

You can login with your AccountName and a PrivateKey which can be generated in the TransIP control panel. The PrivateKey is a stringified version of the private key given by the API, see the example below, each newline is replaced by "\n".

{% highlight json %}
{
  "transip":{
    "AccountName": "your-account-name"
    "PrivateKey": "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp\nwmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5\n1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh\n3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2\npIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX\nGukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il\nAkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF\nL0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k\nX6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl\nU9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ\n37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=\n-----END RSA PRIVATE KEY-----"
  }
}
{% endhighlight %}

Or you can choose to have an AccessToken as credential. These can be generated in the TransIP control panel and have a limited lifetime


{% highlight json %}
{
  "transip":{
    "AccessToken": "your-transip-personal-access-token"
  }
}
{% endhighlight %}



## Metadata

This provider does not recognize any special metadata fields unique to TransIP.

## Usage

Example javascript:

{% highlight js %}
var TRANSIP = NewDnsProvider("transip", "TRANSIP");

D("example.tld", REG_DNSIMPLE, DnsProvider(TRANSIP),
    A("test","1.2.3.4")
);
{% endhighlight %}

## Activation

TransIP depends on a TransIP personal access token.
