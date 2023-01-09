---
title: Check-Creds subcommand
---

# check-creds

This is a stand-alone utility to help verify entries in `creds.json`.

The command does a trivia operation to verify credentials.  If
successful, a list of zones will be output (which may be an empty list). If the credentials or other problems prevent this operation from executing, the exit code will be non-zero and hopefully verbose error messages will be output.

```text
Syntax:

   dnscontrol check-creds [command options] credkey provider

   --creds value   Provider credentials JSON file (default: "creds.json")
   --out value     Instead of stdout, write to this file

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)
   provider: The name of the provider (second parameter to NewDnsProvider() in dnsconfig.js)
```

Starting in [v3.16](v316.md), "provider" is optional.  If it is omitted (or the placeholder value `-` is used), the `TYPE` specified in `creds.json` will be used instead. A warning will be displayed with advice on how to remain compatible with v4.0.

Starting in v4.0, the "provider" argument is expected to go away.

EXAMPLES:

```shell
dnscontrol check-creds myr53 ROUTE53
```

Starting in [v3.16](v316.md):

```shell
dnscontrol check-creds myr53
dnscontrol check-creds myr53 -
dnscontrol check-creds myr53 ROUTE53
```

Starting in v4.0:

```shell
dnscontrol check-creds myr53
```

This command is the same as `get-zones` with `--format=nameonly`

# Developer Note

This command is not implemented for all providers.

To add this to a provider, implement the get-zones subcommand.
