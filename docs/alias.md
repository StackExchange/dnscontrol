---
layout: default
---

# ALIAS records

ALIAS records are not widely standardized across DNS providers. Some (Route 53, DNSimple) have a native ALIAS record type. Others (Cloudflare) implement transparent CNAME flattening.

DNSControl adds an ALIAS record type, and leaves it up to the provider implementation to handle it.

A few notes:

1. A provider must "opt-in" to supporting ALIAS records. When registering a provider, you specify which capabilities you support. Here is an example of how the 
  cloudflare provider declares its support for aliases:

```
func init() {
	providers.RegisterDomainServiceProviderType("CLOUDFLAREAPI", newCloudflare, providers.CanUseAlias)
}
```

2. If you try to use ALIAS records, **all** dns providers for the domain must support ALIAS records. We do not want to serve inconsistent records across providers.
3. CNAMEs at `@` are disallowed, but ALIAS is allowed. 
4. Cloudflare does not have a native ALIAS type, but CNAMEs behave similarly. The Cloudflare provider "rewrites" ALIAS records to CNAME as it sees them. Other providers may not need this step.

