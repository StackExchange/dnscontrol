# Security Policy

DNSControl is a command-line tool and therefore has a different (limited) attack surface as compared to a web app or other system.

## Supported Versions

Only the most recent release is supported with security updates.

When a major version is incremented, we'll support the previous major version for 6 months. For example, when v4.0 is released, we will support the most recent v3.x release for 6 months.

## Reporting a Vulnerability

To report a vulnerability please [create a new GitHub "issue"](https://github.com/StackExchange/dnscontrol/issues/new/choose).

We will respond in a best-effort manner, usually within 1 week. We will communciate via the GitHub issue unless we need to communicate privately, in which case we'll arrange a way to communicate directly.

## Build Attestation

DNSControl uses GitHub Actions and workflows from the SLSA Framework to produce verifiable builds.

<!-- FIXME: version reference below -->

The [releases page](https://github.com/StackExchange/dnscontrol/releases) includes an attestation document (`multiple.intoto.jsonl`) in the list of files associated with each release since vFIXME. This file contains the signed attestation that can be used to verify the provenance of the files associated with each release.

The [SLSA verifier](https://github.com/slsa-framework/slsa-verifier) tool can be used to confirm the authenticity of DNSControl releases. To manually verify a downloaded build artifact and the `multiple.intoto.jsonl` file, use the `slsa-verifier` utility to confirm the artifact was signed:

```shell
slsa-verifier verify-artifact dnscontrol_4.19.0_darwin_all.tar.gz \
  --provenance-path multiple.intoto.jsonl \
  --source-uri github.com/StackExchange/dnscontrol \
  --source-tag v4.20.0
```
