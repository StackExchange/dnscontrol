---
name: BIND
title: BIND Provider
layout: default
jsId: BIND
---
# BIND Provider
This provider maintains a directory with a collection of .zone files.

This provider does not generate or update the named.conf file, nor does it deploy the .zone files to the BIND master.
Both of those tasks are different at each site, so they are best done by a locally-written script.


## Configuration
In your credentials file (`creds.json`), you can specify a `directory` where the provider will look for and create zone files. The default is the `zones` directory where dnscontrol is run.

{% highlight json %}
{
  "bind": {
    "directory": "myzones"
  }
}
{% endhighlight %}

The BIND provider does not require anything in `creds.json`. It does accept some optional metadata via your DNS config when you create the provider:

{% highlight javascript %}
var BIND = NewDnsProvider('bind', 'BIND', {
        'default_soa': {
        'master': 'ns1.example.tld.',
        'mbox': 'sysadmin.example.tld.',
        'refresh': 3600,
        'retry': 600,
        'expire': 604800,
        'minttl': 1440,
    },
    'default_ns': [
        'ns1.example.tld.',
        'ns2.example.tld.',
        'ns3.example.tld.',
        'ns4.example.tld.'
    ]
})
{% endhighlight %}

If you need to customize your SOA or NS records, you can do so with this setup.
