---
name: getConfiguredDomains
parameters:
  - name
  - modifiers...
---

`getConfiguredDomains` getConfiguredDomains is a helper function that returns the domain names
configured at the time the function is called. Calling this function early or later in
`dnsconfig.js` may return different results. Typical usage is to iterate over all
domains at the end of your configuration file.

Example for adding records to all configured domains:
{% include startExample.html %}
{% highlight js %}
var domains = getConfiguredDomains();
for(i = 0; i < domains.length; i++) {
  D_EXTEND(domains[i],
    TXT('_important', 'BLA') // I know, not really creative.
  )
}
{%endhighlight%}

This will end up in following modifications:
```
******************** Domain: domain1.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...2 corrections
#1: CREATE TXT _important.domain1.tld "BLA" ttl=43200
#2: REFRESH zone domain1.tld

******************** Domain: domain2.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...2 corrections
#1: CREATE TXT _important.domain2.tld "BLA" ttl=43200
#2: REFRESH zone domain2.tld
```
{% include endExample.html %}

Example for adding DMARC report records:
{% include startExample.html %}
This example might be more useful, specially for configuring the DMARC report records. According to DMARC RFC you need to specify `domain2.tld._report.dmarc.domain1.tld` to allow `domain2.tld` to send aggregate/forensic email reports to `domain1.tld`. This can be used to do this in an easy way, without using the wildcard from the RFC.

{% highlight js %}
var domains = getConfiguredDomains();
for(i = 0; i < domains.length; i++) {
	D_EXTEND("domain1.tld",
		TXT(domains[i] + '._report._dmarc', 'v=DMARC1')
	);
}
{%endhighlight%}

This will end up in following modifications:
```
******************** Domain: domain2.tld
----- Getting nameservers from: registrar
----- DNS Provider: registrar...4 corrections
#1: CREATE TXT domain1.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
#2: CREATE TXT domain3.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
#3: CREATE TXT domain4.tld._report._dmarc.domain2.tld "v=DMARC1" ttl=43200
#4: REFRESH zone domain2.tld
```
{% include endExample.html %}
