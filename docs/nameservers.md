---
layout: default
title: Nameservers
---

# Nameservers

DNSControl can handle a variety of provider scenarios for you.

- The same provider is the registrar and DNS server
- Different providers for the registrar and DNS server
- A registrar plus multiple DNS servers
- Additional "shadow" DNS servers (non-authoratative DNS servers,
  often used as backups or as a local cache)

# Examples:

{% include startExample.html %}
{% highlight js %}

// ========== Registrars:

// A normal registrar.
var REG_NAMECOM = NewRegistrar("namedotcom_main", "NAMEDOTCOM");
// The "NONE" registrar is a "fake" registrar that makes no changes.
// This is useful if you don't want DNSControl to control who the
// nameservers are for a domain, or if you use a registrar that doesn't
// offer an API, or if the registrar's API is not implemented in
// DNSControl.
var REG_THIRDPARTY = NewRegistrar("ThirdParty", "NONE");

// ========== DNS Providers:

var DNS_NAMECOM = NewDnsProvider("namedotcom_main", "NAMEDOTCOM");
var DNS_AWS = NewDnsProvider("aws_main", "ROUTE53");
var DNS_GOOGLE = NewDnsProvider("gcp_main", "GCLOUD");
var DNS_CLOUDFLARE = NewDnsProvider("cloudflare_main", "CLOUDFLAREAPI");
var DNS_BIND = NewDnsProvider("bind", "BIND");

// ========== Domains:

// "Keep it simple": Use the same provider as a registrar and DNS service.
// Why? Simplicity.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_NAMECOM),
  A("@", "10.2.3.4")
);

// "Separate DNS server": Use one provider as registrar, a different for DNS service.
// Why? Use any registrar but a preferred DNS provider.
// This is the most common situation.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_AWS),
  A("@", "10.2.3.4")
);

// "Registrar only": Direct the registrar to point to some other DNS provider.
// Why? In this example we're pointing the domain to the nsone.net DNS
// service, which someone else is controlling.
D("example1.com", REG_NAMECOM,
  NAMESERVER("dns1.p03.nsone.net."),
  NAMESERVER("dns2.p03.nsone.net."),
  NAMESERVER("dns3.p03.nsone.net."),
  NAMESERVER("dns4.p03.nsone.net."),
);

// "Custom nameservers": Ignore the provider's default nameservers and substitute our own.
// Why? Rarely used unless the DNS provider's API does not support
// querying what the nameservers are, or the API is returning invalid
// data, or if during initial setup the API returns no information.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_CLOUDFLARE, 0),  // Set the DNS provider but ignore the nameservers it suggests (0 == take zero of the names it reports)
  NAMESERVER("kim.ns.cloudflare.com."),
  NAMESERVER("walt.ns.cloudflare.com."),
  A("@", "10.2.3.4")
);

// "Add additional nameservers." Use the default nameservers from the registrar but add additional ones.
// Why? Usually only to correct a bug or misconfiguration elsewhere.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_NAMECOM),
  NAMESERVER("ns1.myexample.tld"),
  A("@", "10.2.3.4")
);

// "Shadow DNS servers."  Secretly send your DNS records to another server.
// Why? Many possibilities:
/  * You are preparing to move to a different DNS provider and want to test it before you cut over.
/  * You want your DNS records stored somewhere else in case you have to switch over in an emergency.
/  * You are sending the zone to a local caching DNS server.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_NAMECOM), // Our real DNS server
  DnsProvider(DNS_CLOUDFLARE, 0), // Quietly send a copy of the zone here.
  DnsProvider(DNS_BIND, 0), // And here too!
  A("@", "10.2.3.4")
);

// "Zonefile backups." Make backups of the exact DNS records in zone-file format.
// Why? In addition to the usual configuration, write out a BIND-style
// zonefile perhaps for debugging, historical, or auditing purposes.
// NOTE: This won't work if you use pseudo rtypes that BIND doesn't support.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_NAMECOM),
  DnsProvider(DNS_BIND, 0), // Don't activate any nameservers related to BIND.
  A("@", "10.2.3.4")
);

// "Dual DNS Providers": Use two different DNS services:
// Why? Diversity. If one DNS provider goes down, the other will be used.
// Little known fact: Most DNS recursive resolvers monitor which DNS
// servers are performing the best and automatically start avoiding
// the slow or down servers. This means that if you use this technique
// and one DNS provider goes down (like the famous Dyn outage), after a
// while your users won't be affected.  Not all software does this
// properly.
// More info: https://www.dns-oarc.net/files/workshop-201203/OARC-workshop-London-2012-NS-selection.pdf
// NOTE: This is overkill unless you have millions of users and strict up-time requirements.
D("example1.com", REG_NAMECOM,
  DnsProvider(DNS_AWS, 2),  // Take 2 nameservers from AWS
  DnsProvider(DNS_GOOGLE, 2),  // Take 2 nameservers from GCP
  A("@", "10.2.3.4")
);

{%endhighlight%}
{% include endExample.html %}


{% include alert.html text="Note: Not all providers allow full control over the NS records of your zone. It is not recommended to use these providers in complicated scenarios such as hosting across multiple providers. See individual provider docs for more info." %}
