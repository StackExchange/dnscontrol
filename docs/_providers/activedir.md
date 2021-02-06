---
name: ActiveDirectory_PS
layout: default
jsId: ACTIVEDIRECTORY_PS
title: ActiveDirectory_PS Provider
---

# WARNING:

WARNING: This provider is deprecated and will eventually be removed.
Please switch to MSDNS. It is more modern and reliable.  The
`creds.json` fields changed names; otherwise it should be an
uneventful upgrade.

# ActiveDirectory_PS Provider
This provider updates an Microsoft Active Directory server DNS server. It interacts with AD via PowerShell commands that are generated and executed on the local machine. This means that DNSControl must be run on a Windows host. This driver automatically deactivates itself when run on non-Windows systems.

# Running on Non-Windows systems
For debugging and testing on non-Windows systems, a "fake PowerShell" mode can be used, which will activate the driver and simulate PowerShell as follows:

- **Zone Input**: Normally when DNSControl needs to know the contents of an existing DNS zone, it generates a PowerShell command to gather such information and saves a copy in a file called `adzonedump.ZONE.json` (where "ZONE" is replaced with the zone name).  When "fake PowerShell" mode is enabled, the PowerShell command is not run, but the `adzonedump.ZONE.json` file is read. You must generate this file ahead of time (often on a different machine, one that runs PowerShell).
- **Zone Changes**: Normally when DNSControl needs to change DNS records, it executes PowerShell commands as required.  When "fake PowerShell" mode is enabled, these commands are simply logged to a file `dns_update_commands.ps1` and the system assumes they executed.

To activate this mode, set `"fakeps":"true"` inside your credentials file for the provider.

## Configuration

The `ActiveDirectory_PS` provider reads an `ADServer` setting from `creds.json` to know the name of the ActiveDirectory DNS Server to update.

{% highlight javascript %}
{
  "activedir": {
    "ADServer": "ny-dc01"
  }
}
{% endhighlight %}


If you want to modify the "fake powershell" mode details, you can set them in the credentials file:

{% highlight javascript %}
{
  "activedir": {
    "ADServer": "ny-dc01",
    "fakeps": "true",
    "pslog": "powershell.log",
    "psout": "commandsToRun.ps1"
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

To generate a `adzonedump.ZONE.json` file, run `dnscontrol preview` on a Windows system then copy the appropriate file to the system you'll use in "fake powershell" mode.

The `adzonedump.ZONE.json` files should be UTF-16LE encoded. If you hand-craft such a file on a non-Windows system, you may need to convert it from UTF-8 to UTF-16LE using:

    iconv -f UTF8  -t UTF-16LE <adzonedump.FOO.json.utf0 > adzonedump.FOO.json

If you check these files into Git, you should mark them as "binary" in `.gitattributes`.
