---
layout: default
title: Get-Zones subcommand
---

# get-zones (was "convertzone")

DNSControl has a stand-alone utility that will contact a provider,
download the records of one or more zones, and output them to a file
in a variety of formats.

The original purpose of this command is to help convert legacy domains
to DNScontrol (bootstrapping).  Since bootstrapping can not depend on
`dnsconfig.js`, `get-zones` relies on command line parameters and
`creds.json` exclusively.

Syntax:

   dnscontrol get-zones [command options] credkey provider zone [...]

   --creds value   Provider credentials JSON file (default: "creds.json")
   --format value  Output format: dsl tsv pretty (default: "pretty")
   --out value     Instead of stdout, write to this file

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)
   zone:     One or more zones (domains) to download; or "all".

EXAMPLES:
   dnscontrol get-zones myr53 ROUTE53 example.com
   dnscontrol get-zones gmain GANDI_V5 example.comn other.com
   dnscontrol get-zones cfmain CLOUDFLAREAPI all
   dnscontrol get-zones -format=tsv bind BIND example.com
   dnscontrol get-zones -format=dsl -out=draft.js glcoud GCLOUD example.com`,


# Example commands

dnscontrol get-zone

# Developer Note

This command is not implemented for all providers.

To add this to a provider:

1. Document the feature

In the `*Provider.go` file, change the setting to implemented.

* OLD: `  providers.CanGetZones:     providers.Unimplemented(),`
* NEW: `  providers.CanGetZones:     providers.Can(),`

2. Update the docs

```
go generate
```

3. Implement the `GetZoneRecords` function

Find the `GetZoneRecords` function in the `*Provider.go` file.

If currently returns `fmt.Errorf("not implemented")`.

Instead, it should gather the records from the provider
and return them as a list of RecordConfig structs.

The code to do that already exists in `GetDomainCorrections`.
You should extract it into its own function (`GetZoneRecords`), rather
than having it be burried in the middle of `GetDomainCorrections`.
`GetDomainCorrections` should call `GetZoneRecords`.

Once that is done the `get-zone` subcommand should work.

4. Optionally implemement the `ListZones` function

If the `ListZones` function is implemented, the command will activate
the ability to specify `all` as the zone, at which point all zones
will be downloaded.

(Technically what is happening is by implementing the `ListZones`
function, you are completing the `ZoneLister` interface for that
provider.)
