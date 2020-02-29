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
   --format value  Output format: dsl pretty tsv nameonly (default: "pretty")
   --out value     Instead of stdout, write to this file
   --ttl value     Default TTL (0 picks the zone's most common TTL) (default: 0)

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)
   zone:     One or more zones (domains) to download; or "all".

FORMATS:
   --format=dsl      dnsconfig.js format (not perfect, but a decent first draft)
   --format=nameonly Just print the zone names
   --format=pretty   BIND Zonefile format
   --format=tsv      TAB separated value (useful for AWK)

When using `tsv`, the columns are:
   FQDN (the label with the domain)
   ShortName (just the label, "@" if it is the naked domain)
   TTL
   Record Type (A, AAAA, CNAME, etc.)
   Target and arguments (quoted like in a zonefile)

The --ttl flag applies to pretty and dsl formats.

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

**Step 1. Document the feature**

In the `*Provider.go` file, change the setting to implemented.

* OLD: `  providers.CanGetZones:     providers.Unimplemented(),`
* NEW: `  providers.CanGetZones:     providers.Can(),`

**Step 2. Update the docs**

```
go generate
```

**Step 3. Implement the `GetZoneRecords` function**

Find the `GetZoneRecords` function in the `*Provider.go` file.

It currently returns `fmt.Errorf("not implemented")`.

Instead, it should gather the records from the provider
and return them as a list of RecordConfig structs.

The code to do that already exists in `GetDomainCorrections`.
You should extract it into its own function (`GetZoneRecords`), rather
than having it be burried in the middle of `GetDomainCorrections`.
`GetDomainCorrections` should call `GetZoneRecords`.

Once that is done the `get-zone` subcommand should work.

**Step 4. Optionally implemement the `ListZones` function**

If the `ListZones` function is implemented, the "all" special case
will be activated.  In this case, listing a single zone named `all`
will query the provider for the list of zones.

(Technically what is happening is by implementing the `ListZones`
function, you are completing the `ZoneLister` interface for that
provider.)

Implementing the `ListZones` function also activates the `check-creds`
subcommand for that provider. Please add to the provider documentation
a list of error messages that people might see if the credentials are
invalid.  See `docs/_providers/gcloud.md` for examples.
