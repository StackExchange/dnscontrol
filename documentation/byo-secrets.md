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

GitHub Actions has a secure
[secrets storage system](https://docs.github.com/en/free-pro-team@latest/actions/reference/encrypted-secrets).
Those secrets are available to GitHub Actions and are required for the
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

# How it works

Tests are executed if `*_DOMAIN` exists where `*` is the name of the provider.  If the value is empty or
unset, the test is skipped.
For example, if a provider is called `FANCYDNS`, there must
be a secret called `FANCYDNS_DOMAIN`.

# Bring your own secrets

This section describes how to add a provider to the testing system.

Step 1: Create a branch

Create a branch as you normally would to submit a PR to the project.

Step 2: Update `pr_test.yml`

In this branch, edit `.github/workflows/pr_test.yml`:

1. In the `integration-test-providers` section, the name of the provider.

Add your provider's name (alphabetically).
The line looks something like:

{% code title=".github/workflows/pr_test.yml" %}
```
        PROVIDERS: "['BIND','HEXONET','AZURE_DNS','CLOUDFLAREAPI','GCLOUD','NAMEDOTCOM','ROUTE53','CLOUDNS','DIGITALOCEAN','GANDI_V5','HEDNS','INWX','NS1','POWERDNS','TRANSIP']"
```
{% endcode %}

2. Add your providers `_DOMAIN` env variable:

Add it to the `env` section of `integration-tests`.

For example, the entry for BIND looks like:

{% code title=".github/workflows/pr_test.yml" %}
```
        BIND_DOMAIN: ${{ vars.BIND_DOMAIN }}
```
{% endcode %}

3. Add your providers other ENV variables:

If there are other env variables (for example, for an API key), add that as a "secret".

For example, the entry for CLOUDFLAREAPI looks like this:

{% code title=".github/workflows/pr_test.yml" %}
```
        CLOUDFLAREAPI_ACCOUNTID: ${{ secrets.CLOUDFLAREAPI_ACCOUNTID }}
        CLOUDFLAREAPI_TOKEN: ${{ secrets.CLOUDFLAREAPI_TOKEN }}
```
{% endcode %}

Step 3. Submit this PR like any other.


# Caveats

Sadly there is no locking to prevent two PRs from running the same
test on the same domain at the same time.  When that happens, both PRs
running the tests fail. In the future we hope to add some locking.

Also, maintaining a fork requires keeping it up to date. That's a bit
more Git knowledge than I can describe here.  (I'm not a Git expert by
any stretch of the imagination!)
