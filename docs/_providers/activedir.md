---
name: ActiveDirectory_PS
layout: default
jsId: ACTIVEDIRECTORY_PS
title: ActiveDirectory_PS Provider
---
# ActiveDirectory_PS Provider

This provider updates an Microsoft DNS server.

It interacts with the server via PowerShell commands. As a result, DNSControl
must be run on Windows and will automatically disable itself when run on
non-Windows systems.

DNSControl will use `New-PSSession` to execute the commands remotely if
`computername` is set in `creds.json` (see below).

# Running on Non-Windows systems

Currently this driver disables itself when run on Non-Windows systems.

It should be possible for non-Windows hosts with PowerShell Core installed to
execeute commands remotely via SSH. The module used to talk to PowerShell
supports this. It should be easy to implement. Volunteers requested.

## Configuration

The `ActiveDirectory_PS` provider reads an `computername` setting from
`creds.json` to know the name of the ActiveDirectory DNS Server to run the commands on.
Otherwise

{% highlight javascript %}
{
  "activedir": {
    "dnsserver": "ny-dc01",
    "pssession": "mywindowshost"
  }
}
{% endhighlight %}

An example DNS configuration:

{% highlight javascript %}
var REG_NONE = NewRegistrar('none', 'NONE')
var ACTIVEDIRECTORY = NewDnsProvider("activedir", "ACTIVEDIRECTORY_PS");

D('example.tld', REG_NONE, DnsProvider(ACTIVEDIRECTORY),
      A("test","1.2.3.4")
)
{% endhighlight %}
