---
layout: default
title: Bug Triage Process
---

# Who to assign bugs to?

If an issue is related to a particular provider, assign it to
the person responsible for the provider, as listed in
[Providers]({{site.github.url}}/provider-list)'s "Maintainers of
contributed providers".

Otherwise leave it unassigned until someone grabs it.


# How bugs are classified

labels:

* enhancement: New feature of improvement of existing feature
* bug: feature works wrong or not as expected

priority:

* maybe someday: Low priority

# How to handle a provider request


1. Change the subject to be "Provider request: name of the provider"
1. Set the label `provider-request`
1. Respond to the issue with the message below
1. Close the issue

The [Providers]({{site.github.url}}/provider-list) page is generated
automatically from all the issues tagged `provider-request`:

1. "Requested providers: state=closed, tagged `provider-request`
1. "In progress providers": state=open, tagged `provider-request`, NOT tagged `has-pr`
1. "Providers with open PRs": state=open, tagged `provider-request` AND `has-pr`

Message to requester:

```
Thank you for requesting this provider!

I've tagged this issue as a provider-request.  It will (soon) be listed as a "requested provider" on the provider list web page:
https://stackexchange.github.io/dnscontrol/provider-list

I will now close the issue.  I know that's a bit confusing, but it will remain on the "requested provider" list.

If someone would like to volunteer to implement this, please re-open this issue and add the tag `has-pr`.

We encourage you to try adding this provider yourself.  We've tried to
make the process as friendly as possible.  Many people have reported
that adding a provider was their first experience writing Go.  The
process is documented here:
https://stackexchange.github.io/dnscontrol/writing-providers
If you need assistance, please speak up in this issue and someone will get back to you ASAP.
```
