### Active Directory

This provider updates a DNS Zone in an Active Directory Integrated Zone.

When run on Windows, AD is updated directly. The code generates
PowerShell commands, executes them, and checks the results.
It leaves behind a log file of the commands that were generated.

When run on non-Windows, AD isn't updated because we can't execute
PowerShell at this time.  Instead of reading the existing zone data
from AD, It learns what
records are in the zone by reading
`adzonedump.{ZONENAME}.json`, a file that must be created beforehand.
It does not actually update AD, it generates a file with PowerShell
commands that would do the updates, which you must execute afterwords.
If the `adzonedump.{ZONENAME}.json` does not exist, the zone is quietly skipped.

Not implemented:

* Delete records.  This provider will not delete any records. It will only add
and change existing records. See "Note to future devs" below.
* Update TTLs.  It ignores TTLs.


## required creds.json config

No "creds.json" configuration is expected.

## example dns config js:

```
var REG_NONE = NewRegistrar('none', 'NONE')
var DSP_ACTIVEDIRECTORY_DS = NewDSP("activedir", "ACTIVEDIRECTORY_PS");

D('ds.stackexchange.com', REG_NONE,
    DSP_ACTIVEDIRECTORY_DS,
)


    // records handled by another provider...
);
```

## Special Windows stuff

This provider needs to do 2 things:

* Get a list of zone records:
  * powerShellDump: Runs a PS command that dumps the zone to JSON.
  * readZoneDump: Opens a adzonedump.$DOMAINNAME.json file and reads JSON out of it.  If the file does not exist, this is considered an error and processing stops.

* Update records:
  * powerShellExec: Execute PS commands that do the update.
  * powerShellRecord: Record the PS command that can be run later to do the updates.  This file is -psout=dns_update_commands.ps1

So what happens when?  Well, that's complex.  We want both Windows and Linux to be able to use -fakewindows
for either debugging or (on Windows) actual use.  However only Windows permits -fakewinows=false and actually executes
the PS code.  Here's which algorithm is used for each case:

  * If -fakewindows is used on any system: readZoneDump and powerShellRecord is used.
  * On Windows (without -fakewindows): powerShellDump and powerShellExec is used.
  * On Linux (wihtout -fakewindows): the provider loads as "NONE" and nothing happens.


## Note to future devs

### Why doesn't this provider delete records?

Because at this time Stack doesn't fully control AD zones
using dnscontrol. It only needs to add/change records.

What should we do when it does need to delete them?

Currently NO_PURGE is a no-op.  I would change it to update
domain metadata to flag that deletes should be enabled/disabled.
Then generate the deletes only if this flag exists.  To be paranoid,
the func that does the deleting could check this flag to make sure
that it really should be deleting something.
