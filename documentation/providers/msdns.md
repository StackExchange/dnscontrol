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
* This is a new provider and has not been tested extensively,
  especially the `pssession` feature.

# Running on Non-Windows systems

Currently this driver disables itself when run on Non-Windows systems.

It should be possible for non-Windows hosts with PowerShell Core installed to
execute commands remotely via SSH. The module used to talk to PowerShell
supports this. It should be easy to implement. Volunteers requested.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `MSDNS`
along with other settings:

* `dnsserver`: (optional) the name of the Microsoft DNS Server to communicate with.
* `pssession`: (optional) the name of the PowerShell PSSession host to run commands on.

Example:

```json
{
  "msdns": {
    "TYPE": "MSDNS",
    "dnsserver": "ny-dc01",
    "pssession": "mywindowshost"
  }
}
```

An example DNS configuration:

```javascript
var REG_NONE = NewRegistrar("none");
var DSP_MSDNS = NewDnsProvider("msdns");

D("example.tld", REG_NONE, DnsProvider(DSP_MSDNS),
      A("test", "1.2.3.4")
)
```
