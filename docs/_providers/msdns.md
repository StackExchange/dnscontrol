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
  "msdns": {
    "dnsserver": "ny-dc01",
    "pssession": "mywindowshost"
  }
}
{% endhighlight %}

An example DNS configuration:

{% highlight javascript %}
var REG_NONE = NewRegistrar('none', 'NONE')
var MSDNS = NewDnsProvider("msdns", "MSDNS");

D('example.tld', REG_NONE, DnsProvider(MSDNS),
      A("test","1.2.3.4")
)
{% endhighlight %}


# Converting from `ACTIVEDIRECTORY_PS`

If you were using the `ACTIVEDIRECTORY_PS` provider and are switching to `MSDNS`, make the following changes:

1. In `dnsconfig.js`, change `ACTIVEDIRECTORY_PS` to `MSDNS` in any `NewDnsProvider()` calls.

2. In `creds.json`: Since unused fields are quietly ignored, it is
   safe to list both the old and new options:
  a. Add a field "dnsserver" with the DNS server's name.  (OPTIONAL if dnscontrol is run on the DNS server.)
  b. If the PowerShell commands need to be run on a different host using a `PSSession`, add `pssession: "remoteserver",` where `remoteserver` is the name of the server where the PowerShell commands should run.
  c. The MSDNS provider will quietly ignore `fakeps`, `pslog` and `psout`. Feel free to leave them in `creds.json` until you are sure you aren't going back to the old provider.

During the transition your `creds.json` file might look like:

{% highlight javascript %}
{
  "msdns": {
    "ADServer": "ny-dc01",         << Delete these after you have
    "fakeps": "true",              << verified that MSDNS works
    "pslog": "log.txt",            << properly.
    "psout": "out.txt",
    "dnsserver": "ny-dc01",
    "pssession": "mywindowshost"
  }
}
{% endhighlight %}

3. Run `dnscontrol preview` to make sure the provider works as expected.

4. If for any reason you need to revert, simply change `dnsconfig.js` to refer to `ACTIVEDIRECTORY_PS` again (or use `git` commands).  If you are reverting because you found a bug, please [file an issue](https://github.com/StackExchange/dnscontrol/issues/new).

5. Once you are confident in the new provider, remove `ADServer`, `fakeps`, `pslog`, `psout` from `creds.json`.
