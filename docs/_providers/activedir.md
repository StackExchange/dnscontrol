---
name: ActiveDirectory_PS
layout: default
jsId: ACTIVEDIRECTORY_PS
---
# ActiveDirectory_PS Provider

This provider updates an Microsoft ActiceDirectory server DNS server. It interacts
with AD via PowerShell commands that are generated and executed on the local machine.
This means that DNSControl must be run on a Windows host.
This driver automatically deactivates itself when run on non-Windows systems.

# Running on Non-Windows systems

For debugging and testing on non-Windows systems,
a "fake powershel" mode can be used, which will activate the driver and
simulate PowersShell as follows:

* Zone Input: Normally when DNSControl needs to know the contents
of an existing DNS zone, it generates a PowerShell command to gather
such information and saves a copy in a file called `adzonedump.ZONE.json`
(where "ZONE" is replaced with the zone name).  When "fake ps" mode is enabled,
the PowerShell command is not run, but the `adzonedump.ZONE.json` file is
read. You can generate this file on a Windows system.
* Zone Changes: Normally when DNSControl needs to change DNS records, it
executes PowerShell commands as required.  When "fake ps" mode is enabled, these
commands are simply logged to a file `dns_update_commands.ps1`.

To activate this mode, set `"fakeps":"true"` inside your credentials file for the provider.

## Configuration

The `ActiveDirectory_PS` provider reads an `ADServer` setting from
`creds.json` to know the name of the ActiceDirectory DNS Server to
update.  creds.json:

{% highlight javascript %}
{
  "activedir": {
    "ADServer": "ny-dc01"
  }
}
{% endhighlight %}

If you want to modify the "fake powershell" mode details, you can set them in the credentials file like so:
{% highlight javascript %}
{
  "activedir": {
    "ADServer": "ny-dc01",
    "fakeps": "true",
    "pslog": "powershell.log",
    "psout: "commandsToRun.ps1"
  }
}
{% endhighlight %}

Here is a simple dns configuration. dnsconfig.js:

{% highlight javascript %}
var REG_NONE = NewRegistrar('none', 'NONE')
var DSP_ACTIVEDIRECTORY_DS = NewDnsProvider("activedir", "ACTIVEDIRECTORY_PS");

D('ds.stackexchange.com', REG_NONE, DnsProvider(DSP_ACTIVEDIRECTORY_DS),
      A("api","172.30.20.100")
)
{% endhighlight %}

To generate a `adzonedump.ZONE.json` file, run `dnscontrol preview`
on a Windows system then copy the appropriate file to the system
you'll use in "fake powershell" mode.

The `adzonedump.ZONE.json` files should be UTF-16LE encoded. If you
hand-craft such a file on a non-Windows system, you may need to
convert it from UTF-8 to UTF-16LE using:

    iconv -f UTF8  -t UTF-16LE <adzonedump.FOO.json.utf0 > adzonedump.FOO.json
