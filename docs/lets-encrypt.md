---
layout: default
title: Let's Encrypt Certificate generation
---

# Let's Encrypt Certificate generation

The `dnscontrol get-certs` command will obtain or renew TLS certificates for your managed domains.

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

`get-certs` will attempt to issue any certificates referenced by this file, and will renew or re-issue if the certificate we already have is
close to expiry or if the set of subject names changes for a cert.

## Working directory layout
The `get-certs` command is designed to be run from a working directory that contains all of the data we need,
and stores all of the certificates and other data we generate.

You may store this directory in source control or wherever you like. At Stack Overflow we have a dedicated repository for
certificates, but we take care to always encrypt any private keys with [black box](https://github.com/StackExchange/blackbox) before committing.

The working directory should generally contain:

- `certificates` folder for storing all obtained certificates.
- `.letsencrypt` folder for storing Let's Encrypt account keys, registrations, and other metadata.
- `certs.json` to describe what certificates to issue.
- `dnsconfig.js` and `creds.json` are the main files for other dnscontrol commands.

```
┏━━.letsencrypt
┃  ┗━(let's encrypt account keys and metadata)
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