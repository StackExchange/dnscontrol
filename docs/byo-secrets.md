---
layout: default
title: Bring-Your-Own-Secrets for automated testing
---

# Bring-Your-Own-Secrets for automated testing

Goal: Enable automated integration testing without accidentally
leaking our API keys and other secrets; at the same time permit anyone
to automate their own tests without having to share their API keys and
secrets.

* PR from a project member:
  * Automated tests run for a long list of providers. All officially supported
    providers have automated tests, plus a few others too.
* PR from an external person
  * Automated tests run for a short list of providers. Any test that
    requires secrets are skipped.
* PR from an external person that wants automated tests for their
  provider.
  * They can set up secrets in their own GitHub account for any tests
    they'd like to automate without sharing their secrets.
  * Note: These tests can always be run outside of GitHub at the
    command line.

# Background: How GitHub Actions protects secrets

Github Actions has a secure
[secrets storage system](https://docs.github.com/en/free-pro-team@latest/actions/reference/encrypted-secrets).
Those secrets are available to Github Actions and are required for the
integration tests to communicate with the various DNS providers that
DNSControl supports.

For security reasons, those secrets are unavailable if the PR comes
from outside the project (a forked repo).  This is a good thing.  If
it didn't work that way, a third-party could write a PR that leaks the
secrets without the owners of the project knowing.

The docs (and many blog posts) describe this as forked repos don't
have access to secrets, and instead receive null strings. That's not
actually what's happening.

Actually what happens is the secrets come from the forked repo.  Or,
more precicely, the secrets offerd to a PR come from the repo that the
PR came from.  A PR from DNSControl's owners gets secrets from
[github.com/StackExchange/dnscontrol's secret store](https://github.com/StackExchange/dnscontrol/settings/secrets/actions)
but a PR from a fork, such as
[https://github.com/TomOnTime/dnscontrol](https://github.com/TomOnTime/dnscontrol)
gets its secrets from TomOnTime's secrets.

Our automated integration tests leverages this info to have tests
only run if they have access to the secrets they will need.

# How it works:

Tests are executed if `*_DOMAIN` exists.  If the value is empty or
unset, the test is skipped.  If a test doesn't require secrets, the
`*_DOMAIN` variable is hardcoded.  Otherwise, it is set by looking up
the secret. For example, if a provider is called `JOEDNS`, there must
be a secret called `JOEDNS_DOMAIN`.
