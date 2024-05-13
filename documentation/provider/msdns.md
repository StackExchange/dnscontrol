This provider updates a Microsoft DNS server.

It interacts with the server via PowerShell commands. As a result, DNSControl
must be run on Windows and will automatically disable itself when run on
non-Windows systems.

DNSControl will use `New-PSSession` to execute the commands remotely if
`computername` is set in `creds.json` (see below).

# Caveats

* Two systems updating a zone is never a good idea. If Windows Dynamic
  DNS and DNSControl are both updating a zone, there will be
  unhappiness.  DNSControl will blindly remove the dynamic records
  unless precautions such as `IGNORE*` and `NO_PURGE` are in use.

# Running on Non-Windows systems

Currently this driver disables itself when run on Non-Windows systems.

It should be possible for non-Windows hosts with PowerShell Core installed to
execute commands remotely via SSH. The module used to talk to PowerShell
supports this. It should be easy to implement. Volunteers requested.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `MSDNS`
along with other settings:

* `dnsserver`: (optional) the name of the Microsoft DNS Server to communicate with.
* `psusername`: (optional) the username to connect to the PowerShell PSSession host.
* `pspassword`: (optional) the password to connect to the PowerShell PSSession host.

Example:

{% code title="creds.json" %}
```json
{
  "msdns": {
    "TYPE": "MSDNS",
    "dnsserver": "ny-dc01",
    "psusername": "mywindowsusername",
    "pspassword": "mysupersecurepassword"
  }
}
```
{% endcode %}

An example DNS configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_MSDNS = NewDnsProvider("msdns");

D("example.com", REG_NONE, DnsProvider(DSP_MSDNS),
      A("test", "1.2.3.4"),
END)
```
{% endcode %}
