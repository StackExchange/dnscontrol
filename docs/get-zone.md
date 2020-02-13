---
layout: default
title: Get-Zone subcommand
---

# get-zone (was "convertzone")

DNSControl has a stand-alone utility that will contact a provider,
download the records of a zone, and output them to a file in a variety
of formats.  The purpose of this command is to help convert legacy
domains to DNScontrol (bootstrapping).  Since bootstrapping can not
depend on `dnsconfig.js`, `get-zone` relies on command line parameters
and `creds.json` exclusively.

Syntax:

    `dnscontrol get-zone [command options] credkey provider zone`


   --creds value   Provider credentials JSON file (default: "creds.json")
   --format value  Output format: dsl tsv pretty (default: "pretty")
   --out value     Instead of stdout, write to this file

   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)
   zone:     The name of the zone (domain) to download

EXAMPLES:

   dnscontrol get-zone myr53 ROUTE53 example.com
   dnscontrol get-zone -format=tsv bind BIND example.com
   dnscontrol get-zone -format=dsl -out=draft.js glcoud GCLOUD example.com


# Example commands

dnscontrol get-zone

# Developer Note

This command is not implemented for all providers.

To add this to a provider:

1. Document the feature

In the `*Provider.go` file, change the setting to implemented.

* OLD: `  providers.CanGetZone:     providers.Unimplemented(),`
* NEW: `  providers.CanGetZone:     providers.Can(),`

2. Update the docs

Run

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
