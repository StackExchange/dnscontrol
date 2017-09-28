---
layout: default
---

<div class="row jumbotron">
	<div class="col-md-12">
		<div>
			<h1 class="hometitle">DnsControl</h1>
			<p class="lead">DnsControl is a platform for seamlessly managing your DNS configuration across any number of DNS hosts, both in the cloud or in your own infrastructure. It manages all of the domains for the Stack Exchange network.</p>
		</div>
	</div>
</div>

<div class="row text-center" style="padding-top: 75px;">
	<div class="col-md-4">
		<h3>Try It</h3>
		<p>Want to jump right in? Follow our <strong><a href="getting-started">quick start tutorial</a></strong> on a new domain or <strong><a href="migrating">migrate</a></strong> an existing one.</p>
	</div>
	<div class="col-md-4">
		<h3>Download It</h3>
		<p>Download the prebuilt binaries for <strong><a href="https://github.com/StackExchange/dnscontrol/releases">binaries</a></strong> and our optional but valuable monitoring agent (Currently works only with OpenTSDB) <strong><a href="/scollector">scollector</a></strong> for Windows, Linux, and Mac.</p>
	</div>
	<div class="col-md-4">
		<h3>Get Help</h3>
		<p>Join us in our Slack room. <a href="/slackInvite">Get an invite.</a> You can <strong><a href="https://github.com/bosun-monitor/bosun/issues">open issues on GitHub</a></strong> to report bugs or discuss new features.</p>
	</div>
</div>

<div class="row" style="padding-top: 75px"><div class='col-md-4 col-md-offset-4'><h2 class="text-center feature-header">Features</h2></div></div>
<hr class="feature">

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/sound-mute.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Use Bosun's flexible expression language to evaluate time series in an exacting way</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/inbox.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Create notifications using Bosun's template language: include graphs, tables, and contextual information</p>
		</div>
	</div>
</div>
<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/sound-mute.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Use Bosun's flexible expression language to evaluate time series in an exacting way</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/inbox.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Create notifications using Bosun's template language: include graphs, tables, and contextual information</p>
		</div>
	</div>
</div>

## Getting Started

### [Getting Started]({{site.github.url}}/getting-started): A walk-through of the basics.

### [Providers]({{site.github.url}}/provider-list): Which DNS providers are supported.

### [Examples]({{site.github.url}}/examples): The DNSControl language by example.

### [Migrating]({{site.github.url}}/migrating): Migrating zones to DNSControl.


## Reference

### [Language Reference]({{site.github.url}}/js): Description of the DNSControl language (DSL).

### [ALIAS / ANAME records in dnscontrol]({{site.github.url}}/alias)

### [Why CNAME/MX/NS targets require a trailing "dot"]({{site.github.url}}/why-the-dot)


## Advanced Usage

### [Testing]({{site.github.url}}/unittests): Unit Testing for you DNS Data.

## Developer info

### [github](https://github.com/StackExchange/dnscontrol): Get the source!

### [Writing Providers]({{site.github.url}}/writing-providers)

### [Adding new DNS record types]({{site.github.url}}/adding-new-rtypes)

