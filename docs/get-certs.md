---
layout: default
title: Let's Encrypt Certificate generation
---

# *Let's Encrypt* Certificate generation

DNSControl will generate/renew Let's Encrypt certificates using DNS
validation.  It is not a complete certificate management system, but
can perform the renewal steps for the system you create.  If you
are looking for a complete system, we recommend
[certbot](https://certbot.eff.org/).

The `dnscontrol get-certs` command will obtain or renew TLS
certificates for your managed domains via
[*Let's Encrypt*](https://letsencrypt.org). This can be extremely useful in
situations where other acme clients are problematic. Specifically,
this may be useful if:

- You are already managing DNS records with DNSControl.
- You have a large number of domains or DNS providers in complicated configurations.
- You want **wildcard** certificates, which *require* DNS validation.

At Stack Overflow we have dual-hosted DNS i.e. zones having
nameservers at two different DNS providers. Most Let's Encrypt systems
do not support DNS validation in that case.  DNSControl's `get-certs`
command leverages the core DNSControl commands when issueing
certificates, therefore dual-hosted DNS is supported.

## General Process

The `get-certs` command does the following steps:

1. Determine which certificates you would like issued, and which names should belong to each one.
1. Look for existing certs on disk, and see if they have sufficient time remaining until expiration, and that the names match.
1. If updates are needed:
    1. Request a new certificate from the acme server.
    1. Receive a list of validations to fill.
    1. For each validation (usually one per name on the cert):
        1. Create a TXT record on the domain with a given secret value.
        1. Wait until the authoritative name servers all return the correct value (polls locally).
        1. Tell the acme server to validate the record.
    1. Receive a new certificate and save it to disk

Because DNS propagation times vary from provider to provider, and
validations are (currently) done serially, this process may take some
time.

## certs.json

This file should be provided to specify which names you would like to get certificates for. You can
specify any number of certificates, with up to 100 SAN entries each. Subject names can contain wildcards if you wish.

The format of the file is a simple json array of objects:

```
[
    {
        "cert_name": "mainCert",
        "names": [
            "example.com.com",
            "www.example.com"
        ]
    },
    {
        "cert_name": "wildcardCert",
        "names": [
            "example.com",
            "*.example.com",
            "*.foo.example.com",
            "otherdomain.tld",
            "*.otherdomain.tld"
        ]
    }
]
```

`dnscontrol get-certs` will attempt to issue any certificates referenced by this file, and will renew or re-issue if the certificate we already have is
close to expiry or if the set of subject names changes for a cert.

## Working directory layout
The `get-certs` command is designed to be run from a working directory that contains all of the data we need,
and stores all of the certificates and other data we generate.

You may store this directory in source control or wherever you like. At Stack Overflow we have a dedicated repository for
certificates, but we take care to always encrypt any private keys with [black box](https://github.com/StackExchange/blackbox) before committing.

The working directory should generally contain:

- `certificates` folder for storing all obtained certificates.
- `.letsencrypt` folder for storing *Let's Encrypt* account keys, registrations, and other metadata.
- `certs.json` to describe what certificates to issue.
- `dnsconfig.js` and `creds.json` are the main files for other dnscontrol commands.

```
┏━━.letsencrypt
┃  ┗━(*Let's Encrypt* account keys and metadata)
┃
┣━━certificates
┃  ┣━━mainCert
┃  ┃  ┣━mainCert.crt
┃  ┃  ┣━mainCert.json
┃  ┃  ┗━mainCert.key
┃  ┗━━wildcardCert
┃     ┣━wildcardCert.crt
┃     ┣━wildcardCert.json
┃     ┗━wildcardCert.key
┃
┣━━certs.json
┣━━creds.json
┗━━dnsconfig.js
```
## Command line flags

### Required Flags

- `--email test@example.com`: Email address to use for *Let's Encrypt* account registration.
- `--agreeTOS`: Indicates that you agree to the [*Let's Encrypt* Subscriber Agreement](https://letsencrypt.org/documents/LE-SA-v1.2-November-15-2017.pdf)

### Optional Flags

- `--config {dnsconfig.js}`, `--creds {creds.json}` and other flags to find your dns configuration are the same as used for `dnscontrol preview` or `push`. `get-certs` needs to read the dns config so it knows which providers manage which domains, and so it can make sure it is not going to make any destructive changes to your domains. If the `get-certs` command needs to fill a challenge on a domain that has pending corrections, it will abort for safety. You can run `dnscontrol preview` and `dnscontrol push` at that point to verify and push the pending corrections, and then proceed with issuing certificates.
- `--acme {url}`: URL of the acme server you wish to use. For *Let's Encrypt* you can use the presets `live` or `staging` for the standard services. If you are using a custom boulder instance or other acme server, you may specify the full **directory** url. Must be an acme **v2** server.
- `--renew {n}`: `get-certs` will renew certs with less than this many **days** remaining. The default is 15, and certs will be renewed when they are within 15 days of expiration.
- `--dir {d}`: Root directory holding all certificate and account data as described above. Default is current working directory.
- `--certConfig {j}`: Location of certificate config json file as described above. Default is `./certs.json`
- `--vault` Store certificates as secrets in hashicorp vault instead of on disk. (default: false)
- `--vaultPath {value}` Path in vault to store certificates (default: "/secret/certs")
- `--skip {p}`: DNS Provider names (comma separated) to skip using as challenge providers. We use this to avoid unnecessary changes to our backup or internal dns providers that wouldn't be a part of the validation flow.
- `--notify` set to true to send notifications to configured destinations (default: false)
- `--only {value}` Only check a single cert. Provide cert name.


## Workflow

This command is intended to be just a small part of a full certificate automation workflow. It only issues certificates, and explicitly does not deal with certificate storage or deployment. We urge caution to secure your private keys for your certificates, as well as the *Let's Encrypt* account private key. We use [black box](https://github.com/StackExchange/blackbox) to securely store private keys in the certificate repo.

This command is intended to be run as frequently as you desire. One workflow would be to check all certificates into a git repository and run a nightly build that:

1. Clones the cert repo, and the dns config repo (if separate).
2. Decrypt or otherwise obtain the *Let's Encrypt* account private key. Dnscontrol does not need to read any certificate private keys to check or issue certificates.
3. Run `dnscontrol get-certs` with appropriate flags.
4. Encrypt or store any new or updated private keys.
5. Commit and push any changes to the cert repo.
6. Take care to not leave any plain-text private keys on disk.

The push to the certificate repo can trigger further automation to deploy certs to load balancers, cdns, applications and so forth.

## Example script

```

#!/bin/bash

set -e

# get and decrypt the files
[ insert your own code here ]

dnscontrol get-certs \
-email "CHANGE_THIS@example.com" \
--acme live \
--skip bind --renew 31 \
--verbose \
--agreeTOS --vault --notify

# Encrypt and save the files
[ insert your own code here ]
```
