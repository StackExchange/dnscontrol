# init

`dnscontrol init` walks a new user through creating a working `creds.json`
plus a minimal `dnsconfig.js` starter. It prompts for a DNS provider and a
registrar, shows where to create the matching API credentials, and writes
both files so that `dnscontrol preview` runs on a fresh checkout.

```shell
NAME:
   dnscontrol init - Interactively create a creds.json and starter dnsconfig.js

USAGE:
   dnscontrol init [command options] [arguments...]

CATEGORY:
    main

OPTIONS:
   --creds value     Output path for the credentials file (default: "creds.json")
   --config value    Output path for the starter DNSControl config (default: "dnsconfig.js")
   --no-config       Do not write a starter dnsconfig.js
   --help, -h        show help
```

## What happens

1. Pick a DNS provider. This is the service that hosts the actual records
   (A, MX, TXT, CNAME, and so on). Pick `NONE` to defer the choice.
2. If the chosen DNS provider also works as a registrar, `init` offers to
   reuse the same account for nameserver (NS) delegation.
3. Otherwise pick a registrar. The registrar is where the domain itself
   is registered. Pick `NONE` to manage the registrar outside DNSControl.
4. For each provider `init` prints the API settings URL so you can open
   it from your terminal before answering the prompts.
5. `init` prompts for every `creds.json` field registered for the
   provider. Secret fields mask the input. Fields that carry newlines,
   such as a PEM encoded private key, open your `$EDITOR` so the full
   block can be pasted without being split up. Providers that support
   multiple auth methods (for example TransIP) use an internal selector
   so only the relevant fields are prompted.
6. You name one or more domains for the starter `dnsconfig.js`.
7. Before writing, `init` shows a preview of both files and asks for
   confirmation.
8. After writing, `init` offers to call
   `dnscontrol get-zones --format=nameonly` against the provider and
   lists which configured domains exist at the provider, which are only
   in the config and which are only at the provider.
9. `init` offers to run `dnscontrol preview` as a final sanity check.

Existing `creds.json` entries are preserved when new entries are added.
An existing `dnsconfig.js` is replaced by the starter; if you want to
keep your current file, pass `--no-config` or answer no at the final
confirmation.

## Provider coverage

`init` only lists providers whose maintainers have registered onboarding
metadata. Other providers can still be used with DNSControl; their
`creds.json` entry is created from the provider's documentation page at
`https://docs.dnscontrol.org/provider/` instead. Provider maintainers
who want their provider to appear in the wizard can find the metadata
schema and flag reference in
[Writing new DNS providers](../advanced-features/writing-providers.md).

## Example: Cloudflare (single account for DNS and registrar)

```shell
$ dnscontrol init

A DNS provider hosts the records (A, MX, TXT, CNAME, and so on) for your zones.
Pick NONE if you want to defer this choice.
? Which DNS service provider do you want to configure? CLOUDFLAREAPI
? Use the same Cloudflare account for the registrar role too? Yes

== DNS provider: Cloudflare ==

API settings for Cloudflare: https://dash.cloudflare.com/profile/api-tokens
? API Token (required) **********
? Account ID (optional) 0123456789abcdef
? creds.json entry name for this provider cloudflare_primary
? First domain name for dnsconfig.js example.com
? Add another domain? No
...
? Write these files? Yes

Done.
? Compare domains in dnsconfig.js with zones at Cloudflare? Yes

$ dnscontrol get-zones --format=nameonly -- cloudflare_primary - all
Zones at Cloudflare compared with dnsconfig.js:
  In both          : example.com
  Only in config   : (none)
  Only at provider : other.com

? Run `dnscontrol preview` now? Yes

Welcome to the DNSControl community!
...
```

## Example: TransIP for DNS with a PEM private key

TransIP accepts either a short lived access token, or an account name
combined with a long lived private key. `init` asks which method you
want to use and only prompts for the relevant fields. The private key
field is multi line, so `init` opens your `$EDITOR` where you paste the
full PEM block (including the `BEGIN` and `END` lines). Set `$EDITOR`
before running `init` (for example `export EDITOR=nano`).

```shell
$ dnscontrol init

A DNS provider hosts the records (A, MX, TXT, CNAME, and so on) for your zones.
Pick NONE if you want to defer this choice.
? Which DNS service provider do you want to configure? TRANSIP

== DNS provider: TransIP ==

API settings for TransIP: https://www.transip.nl/cp/account/api/
TransIP supports two auth methods: a short lived access token, or an account
name paired with a long lived private key.
? Which authentication method do you want to use? Account name + private key
? Account name my-account
? Private key (opens $EDITOR) [Enter to launch editor]
...
```
