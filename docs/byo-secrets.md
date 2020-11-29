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
    requires secrets are skipped in the fork. They will run after the fact though
    once the PR has been merged to into the `master` branch of StackExchange/dnscontrol.
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
more precisely, the secrets offered to a PR come from the repo that the
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
the secret. For example, if a provider is called `FANCYDNS`, there must
be a secret called `FANCYDNS_DOMAIN`.

# Bring your own secrets

This section describes how to add a provider to the testing system.

In this example, we will use a fictional DNS provider named
"FANCYDNS".

Step 1: Create a branch

Create a branch as you normally would to submit a PR to the project.

Step 2: Update `build.yml`

In this branch, edit `.github/workflows/build.yml`:

1. In the `integration-tests` section, add the name of your provider
   to the matrix of providers.  Technically you are adding to the list
   at `jobs.integration-tests.strategy.matrix.provider`.

```
      matrix:
        provider:
...
        - DIGITALOCEAN
        - GANDI_V5
        - FANCYDNS          <<< NEW ITEM ADDED HERE
        - INWX
```

2. Add your test's env:

Locate the env section (technically this is `jobs.integration-tests.env`) and
add all the env names that your provider sets in
`integrationTest/providers.json`.

Please replicate the formatting of the existing entries:

* A blank comment separates each provider's section.
* The providers are listed in the same order as the matrix.provider list.
* The `*_DOMAIN` variable is first.
* The remaining variables are sorted lexicographically (what nerds call alphabetical order).

```
      FANCYDNS_DOMAIN: ${{ secrets.FANCYDNS_DOMAIN }}
      FANCYDNS_KEY: ${{ secrets.FANCYDNS_KEY }}
      FANCYDNS_USER: ${{ secrets.FANCYDNS_USER }}
```

# Examples

Let's look at three examples:

## Example 1:

The `BIND` integration tests do not require any secrets because it
simply generates files locally.

```
      BIND_DOMAIN: example.com
```

The existence of `BIND_DOMAIN`, and the fact that the value is
available to all, means these tests will run for everyone.

## Example 2:

The `AZURE_DNS` provider requires many settings. Since
`AZURE_DNS_DOMAIN` comes from GHA's secrets storage, we can be assured
that the tests will skip if the PR does not have access to the
secrets.

If you have a fork and want to automate the testing of `AZURE_DNS`,
simply set the secrets named in `build.yml` and the tests will
activate for your PRs.

Note that `AZURE_DNS_RESOURCE_GROUP` is hardcoded to `DNSControl`. If
this is not true for you, please feel free to submit a PR that turns
it into a secret.

```
      AZURE_DNS_DOMAIN: ${{ secrets.AZURE_DNS_DOMAIN }}
      AZURE_DNS_CLIENT_ID: ${{ secrets.AZURE_DNS_CLIENT_ID }}
      AZURE_DNS_CLIENT_SECRET: ${{ secrets.AZURE_DNS_CLIENT_SECRET }}
      AZURE_DNS_RESOURCE_GROUP: DNSControl
      AZURE_DNS_SUBSCRIPTION_ID: ${{ secrets.AZURE_DNS_SUBSCRIPTION_ID }}
      AZURE_DNS_TENANT_ID: ${{ secrets.AZURE_DNS_TENANT_ID }}
```

## Example 3:

The HEXONET integration tests require secrets, but HEXONET provides an
Operational Test and Evaluation (OT&E) environment with some "fake"
credentials which are known publicly.

Therefore, since there's nothing secret about these particular
secrets, we hard-code them into the `build.yml` file. Since
`HEXONET_DOMAIN` does not come from secret storage, everyone can run
these tests. (We are grateful to HEXONET for this public service!)

```
      HEXONET_DOMAIN: a-b-c-movies.com
      HEXONET_ENTITY: OTE
      HEXONET_PW: test.passw0rd
      HEXONET_UID: test.user
```

NOTE: The above credentials are [known to the public]({{site.github.url}}/providers/hexonet).


# Caveats

Sadly there is no locking to prevent two PRs from running the same
test on the same domain at the same time.  When that happens, both PRs
running the tests fail. In the future we hope to add some locking.

Also, maintaining a fork requires keeping it up to date. That's a bit
more Git knowledge than I can describe here.  (I'm not a Git expert by
any stretch of the imagination!)
