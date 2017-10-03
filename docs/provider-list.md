---
layout: default
---
<h1> Service providers </h1>

<table class='table table-bordered'>
  <thead>
    <th>Name</th>
    <th>Identifier</th>
  </thead>
{% for p in site.providers %}
<tr>
  <td><a href=".{{p.id}}">{{p.name}}</a></td>
  <td>{{p.jsId}}</td>
</tr>
{% endfor %}
</table>

<h2> Provider Features </h2>

<p>The table below shows various features supported, or not supported by DNSControl providers.
  Underlined items have tooltips for more detailed explanation. This table is automatically generated
  from metadata supplied by the provider when they register themselves inside dnscontrol.
</p>
<p>
  An empty space may indicate the feature is not supported by a provider, or it may simply mean
  the feature has not been investigated and implemented yet. If a feature you need is missing from 
  a provider that supports it, we'd love your contribution to ensure it works correctly and add it to this matrix.
</p>
<p>If a feature is definitively not supported for whatever reason, we would also like a PR to clarify why it is not supported, and fill in this entire matrix.</p>
<br/>
<br/>

{% include matrix.html %}


### Providers with "official support"

Official support means:

* New releases will block if any of these providers do not pass integration tests.
* The DNSControl maintainers prioritize fixing bugs in these providers (though we gladly accept PRs).
* New features will work on these providers (unless the provider does not support it).
* StackOverflow maintains test accounts with those providers for running integration tests.

### Providers with "contributor support"

The other providers are supported by community members, usually the
original contributor.

Due to the large number of DNS providers in the world, the DNSControl
team can not support and test all providers.  Test frameworks are
provided to help community members support their code independently.

* Maintainers are expected to support their provider and/or find a new maintainer.
* Bugs will be referred to the original contributor or their designate.
* Maintainers should set up test accounts and regularly verify that all tests pass (`pkg/js/parse_tests` and `integrationTest`).
* Contributors are encouraged to add new tests and refine old ones. (Test-driven development is encouraged.)

Maintainers of contributed providers:

* digital ocean @Deraen
* dnsimple @aeden
* gandi @TomOnTime
* namecheap @captncraig
* ns1 @captncraig
* OVH @Oprax

### Requested providers

We have received requests for the following providers. If you would like to contribute
code to support this provider, please re-open the issue. We'd be glad to help in any way.

<ul>
  <li>Azure (<a href="https://github.com/StackExchange/dnscontrol/issues/42">#42</a>)</li>
  <li>ClouDNS (<a href="https://github.com/StackExchange/dnscontrol/issues/114">#114</a>)</li>
  <li>Dyn (<a href="https://github.com/StackExchange/dnscontrol/issues/61">#61</a>)</li>
  <li>Gandi (DNS works. Request is to add Registrar support) (<a href="https://github.com/StackExchange/dnscontrol/issues/87">#87</a>)</li>
  <li>GoDaddy (<a href="https://github.com/StackExchange/dnscontrol/issues/145">#145</a>)</li>
  <li>Hurricane Electric (dns.he.net) (<a href="https://github.com/StackExchange/dnscontrol/issues/118">#118</a>)</li>
  <li>Linode (<a href="https://github.com/StackExchange/dnscontrol/issues/121">#121</a>)</li>
  <li>OVH (<a href="https://github.com/StackExchange/dnscontrol/issues/143">#143</a>)</li>
</ul>
