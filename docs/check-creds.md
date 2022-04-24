---
layout: default
title: Check-Creds subcommand
---

# check-creds

This is a stand-alone utility to help verify entries in `creds.json`.

The command does a trivia operation to verify credentials.  If
successful, a list of zones will be output. If not, hopefully you see
verbose error messages.

Syntax:

   dnscontrol check-creds [command options] credkey provider

   --creds value   Provider credentials JSON file (default: "creds.json")
   --out value     Instead of stdout, write to this file

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)

If "provider" is `-`, the provider type is assumed to be the value of `TYPE` in `creds.json`.  Prior to v4.0 this parameter will become optional.  In v4.0 this paramenter is expected to go away.

EXAMPLES:
   dnscontrol check-creds myr53 ROUTE53

This command is the same as:
   dnscontrol get-zones --out=/dev/null myr53 ROUTE53

# Developer Note

This command is not implemented for all providers.

To add this to a provider, implement the get-zones subcommand
