---
name: Bind
layout: default
jsId: BIND
---
# Bind Provider

This provider simply maintains a directory with a collection of .zone files. We currently copy zone files to our production servers and restart bind via 
a script external to DNSControl.

## Configuration

The BIND provider does not require anything in `creds.json`. It does accept some (optional) metadata via your dns config when you create the provider:

{% highlight javascript %}
var bind = NewDnsProvider('bind', 'BIND', {
  'default_soa': {
    'master': 'ns1.mydomain.com.',
    'mbox': 'sysadmin.mydomain.com.',
    'refresh': 3600,
    'retry': 600,
    'expire': 604800,
    'minttl': 1440,
  },
  'default_ns': [
        'ns1.mydomain.com.',
        'ns2.mydomain.com.',
        'ns3.mydomain.com.',
        'ns4.mydomain.com.'
  ]
})
{% endhighlight %}

If you need to customize your SOA or NS records, you can do it with this setup.

You can also provide a `-bindtree=directoryName` flag to change where the provider will look for and create zone files. The default is the `zones` directory where dnscontrol is run.
