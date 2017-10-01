---
layout: default
---

<div class="container-fluid">
	<div class="row">
		<div class="col-md-12">
			<div class="page-header">
				<h1>
					DNSControl: <small>DNS as Code</small>
				</h1>
			</div>
		</div>
	</div>
	<div class="row">
		<div class="col-md-4">
			<h2>
				Getting Started
			</h2>
			<p>
				Information for new users and the curious.
			</p>

			<ul>
				<li>
          <a href="{{site.github.url}}/getting-started">Getting Started</a>: A walk-through of the basics
				</li>
				<li>
					<a href="{{site.github.url}}/provider-list">Providers</a>: Which DNS providers are supported
				</li>
				<li>
					<a href="{{site.github.url}}/examples">Examples</a>: The DNSControl language by example
				</li>
				<li>
					<a href="{{site.github.url}}/migrating">Migrating</a>: Migrating zones to DNSControl
				</li>
			</ul>
		</div>
		<div class="col-md-4">
			<h2>
				Reference
			</h2>
			<p>
				Language resources and procedures.
			</p>

			<ul>
				<li>
					<a href="{{site.github.url}}/js">Language Reference</a>: Description of the entire language
				</li>
				<li>
					<a href="{{site.github.url}}/alias">ALIAS / ANAME records in dnscontrol</a>
				</li>
				<li>
					<a href="{{site.github.url}}/spf">SPF Optimizer</a>: Optimize your SPF records
				</li>
			</ul>
		</div>
		<div class="col-md-4">
			<h2>
				Advanced Topics
			</h2>
			<p>
				Take advantage of DNSControl's unique features.
			</p>
			<ul>
				<li>
					<a href="">Why CNAME/MX/NS targets require a trailing "dot{{site.github.url}}/why-the-dot"</a>
				</li>
				<li>
					<a href="{{site.github.url}}/unittests">Testing</a>: Unit Testing for you DNS Data
				</li>

			</ul>
		</div>
	</div>
	<div class="row">
		<div class="col-md-12">
			<h2>
				Developer Info
			</h2>
			<p>
				It is easy to add features and new providers to DNSControl. The code is very modular and easy to modify. There are extensive integration tests that make it easy to boldly make changes with confidence that you'll know if anything is broken. Our mailing list is friendly. Afraid to make your first PR? We'll gladly mentor you through the process. Many major code contributions have come from <a href="https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html">first-time Go users</a>!
			</p>
			<ul>
				<li>
					Github: <a href="https://github.com/StackExchange/dnscontrol">https://github.com/StackExchange/dnscontrol</a>
				</li>
				<li>
					Mailing list: <a href="https://groups.google.com/forum/#!forum/dnscontrol-discuss">dnscontrol-discuss</a>: The friendly best place to ask questions and propose new features
				</li>
				<li>
					<a href="{{site.github.url}}/writing-providers">Step-by-Step Guide: Writing Providers</a>: How to write a DNS or Registrar Provider
				</li>
				<li>
					<a href="{{site.github.url}}/adding-new-rtypes">Step-by-Step Guide: Adding new DNS rtypes</a>: How to add a new DNS record type
				</li>
			</ul>
		</div>
	</div>
</div>
