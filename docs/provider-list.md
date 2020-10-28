---
layout: default
title: Service Providers
---
<h1> Service Providers </h1>

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

<a name="features"></a>
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

Current owners are:

* `ACTIVEDIRECTORY_PS` @tlimoncelli
* `AZURE_DNS` @vatsalyagoel
* `BIND` @tlimoncelli
* `CLOUDFLAREAPI` @tlimoncelli
* `GCLOUD` @tlimoncelli
* `NAMEDOTCOM` @tlimoncelli
* `ROUTE53` @tlimoncelli

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

* `AXFRDDNS` @hnrgrgr
* `CLOUDNS` @pragmaton
* `CSCGLOBAL` @Air-New-Zealand
* `DESEC` @D3luxee
* `DIGITALOCEAN` @Deraen
* `DNSOVERHTTPS` @mikenz
* `DNSIMPLE` @aeden
* `EXOSCALE` @pierre-emmanuelJ
* `GANDI_V5` @TomOnTime
* `HEDNS` @rblenkinsopp
* `HETZNER` @das7pad
* `HEXONET` @papakai
* `INTERNETBS` @pragmaton
* `INWX` @svenpeter42
* `LINODE` @koesie10
* `NAMECHEAP` @captncraig
* `NETCUP` @kordianbruck
* `NS1` @captncraig
* `OCTODNS` @TomOnTime
* `OPENSRS` @pierre-emmanuelJ
* `OVH` @masterzen
* `POWERDNS` @jpbede
* `SOFTLAYER`@jamielennox
* `VULTR` @pgaskin

### Requested providers

We have received requests for the following providers. If you would like to contribute
code to support this provider, please re-open the issue. We'd be glad to help in any way.

<ul id='requests'>

</ul>

### In progress providers

These requests have an *open* issue, which indicates somebody is actively working on it. Feel free to follow the issue, or pitch in if you think you can help.

<ul id='inprog'>
</ul>

### Providers with open PRs

These providers have an open PR with (potentially) working code. They may be ready to merge, or may have blockers. See issue and PR for details.

<ul id='haspr'>
</ul>

<script>
$(function() {
  $.get("https://api.github.com/repos/StackExchange/dnscontrol/issues?state=all&labels=provider-request&direction=asc")
  .done(function(data) {
    for(var i of data) {
      var el = $(`<li><a href='${i.html_url}'>${i.title}</a> (#${i.number})</li>`);
      var target = $("#requests");
      if (i.state == "open") {
        target = $("#inprog");
        for(var l of i.labels) {
          if (l.name == "has-pr")
            target = $("#haspr");
        }
      }
      target.append(el);
    }
  })
  .fail(function(err){
    console.log("???", err)
  });
});
</script>
