---
name: Microsoft DNS Server (Windows Server)
layout: default
jsId: MSDNS
title: Microsoft DNS Server on Microsoft Windows Server
---

# Microsoft DNS Server on Microsoft Windows Server

This provider updates a Microsoft DNS server.

It interacts with the server via PowerShell commands. As a result, DNSControl
must be run on Windows and will automatically disable itself when run on
non-Windows systems.

DNSControl will use `New-PSSession` to execute the commands remotely if
`computername` is set in `creds.json` (see below).

This provider will replace `ACTIVEDIRECTORY_PS` which is deprecated.

# Caveats

* Two systems updating a zone is never a good idea. If Windows Dynamic
  DNS and DNSControl are both updating a zone, there will be
  unhappiness.  DNSControl will blindly remove the dynamic records
  unless precautions such as `IGNORE*` and `NO_PURGE` are in use.
* This is a new provider and has not been tested extensively,
  especially the `pssession` feature.

# Running on Non-Windows systems

Currently this driver disables itself when run on Non-Windows systems.

It should be possible for non-Windows hosts with PowerShell Core installed to
execute commands remotely via SSH. The module used to talk to PowerShell
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
var MSDNS = NewDnsProvider("activedir", "MSDNS");

D('example.tld', REG_NONE, DnsProvider(MSDNS),
      A("test","1.2.3.4")
)
{% endhighlight %}
