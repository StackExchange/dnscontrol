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
