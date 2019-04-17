---
layout: default
title: Nameservers
---

# Nameservers

DNSControl can handle a variety of provider scenarios for you:

- A single provider manages everything for your domains (Ex: name.com registers and serves dns)
- A single provider serves dns seperately from the registrar (Ex: name.com registers and cloudflare hosts dns records)
- Multiple providers "co-host" dns (Ex: Route53 and Google Cloud DNS both serve as authoritative nameservers)
- One or more "active" dns hosts and another "backup" dns host. (Ex: route53 hosts dns, but I update a local bind server as a backup)

All of these scenarios differ in how they manage:

- The root list of authoritative nameservers stored in the tld zone by your registrar.
- The list of NS records for the base domain that is served by each dns host.

DNSControl attempts to manage these records for you as much as possible, according the the following processes:

## 1. Specifying Nameservers

There are several different ways to declare nameservers for a zone:

1. Explicit [`NAMESERVER`](/js#NAMESERVER) records in a domain:

    `NAMESERVER("ns1.myhost.tld")`
2. Request all nameservers to use from a provider (usually via api):

    `DnsProvider(route53)`
3. Request a limited number of nameservers from a provider:

    `DnsProvider(route53, 2), DnsProvider(gcloud, 2)`

    This can be useful to limit the total number of NS records when using multiple providers together, for performance reasons.

## 2. DNSControl processes

The first step in running a domain is a first pass to collect all nameservers that we should use.
DNSControl collects all explicit nameservers, and calls the method on each provider to get nameservers to use.
After this process we have a list of "Authoritative Nameservers" for the domain.

As much as possible, all dns servers should agree on this nameserver list, and serve identical NS records. DNSControl will generate
NS records for the authoritative nameserver list and automatically add them to the domain's records.
NS records for the base domain should not be specified manually, as that will result in an error.

{% include alert.html text="Note: Not all providers allow full control over the NS records of your zone. It is not recommended to use these providers in complicated scenarios such as hosting across multiple providers. See individual provider docs for more info." %}

DnsControl will also register the authoritative nameserver list with the registrar, so that all nameserver are used in the tld registry.

## 3. Backup providers

It is also possible to specify a DNS Provider that is not "authoritative" by using `DnsProvider("name", 0)`. This means the provider will be updated
with all records to match the authoritative ones, but it will not be registered in the tld name servers, and will not take traffic.
It's nameservers will not be added to the authoritative set. While this may seem an attractive option, there are a few things to note:

1. Backup nameservers will still be updated with the NS records from the authoritative nameserver list. This means the records will still need to be updated to correctly "activate" the provider.
2. Costs generally scale with utilization, so there is often no real savings associated with an active-passive setup vs an active-active one anyway.

