---
layout: default
title: DNSControl
---

<div class="row jumbotron">
	<div class="col-md-12">
		<div>
			<h1 class="hometitle">DNSControl</h1>
			<p class="lead">DNSControl is an <strong><a href="opinions">opinionated</a></strong> platform for seamlessly managing your DNS configuration across any number of DNS hosts, both in the cloud or in your own infrastructure. It manages all of the domains for the Stack Overflow network, and can do the same for you!</p>
		</div>
	</div>
</div>

<div class="row text-center" style="padding-top: 75px;">
	<div class="col-md-4">
		<h3>Try It</h3>
		<p>Want to jump right in? Follow our
         <strong><a href="getting-started">quick start tutorial</a></strong>
         on a new domain or
         <strong><a href="migrating">migrate</a></strong>
         an existing one. Read the
         <strong><a href="js">language spec</a></strong>
         for more info. You can also <strong><a href="toc">view a list of all topics</a></strong>.
    </p>
	</div>

	<div class="col-md-4">
		<h3>Use It</h3>
		<p>Take advantage of the
         <strong><a href="">advanced features</a></strong>.
         Use macros and variables for easier updates.
         <!-- Optimize your SPF records. -->
         Upload your zones to
         <strong><a href="provider-list">multiple DNS providers</a></strong>.
    </p>
	</div>

	<div class="col-md-4">
		<h3>Get Involved</h3>
		<p>Join our
         <strong><a href="https://groups.google.com/forum/#!forum/dnscontrol-discuss">mailing list</a></strong>.
         We make it easy to contribute by using
         <strong><a href="https://github.com/StackExchange/dnscontrol">GitHub</a></strong>,
         you can make code changes with confidence thanks to extensive integration tests.
         The project is 
         <strong><a href="https://everythingsysadmin.com/2017/08/go-get-up-to-speed.html">newbie-friendly</a></strong>
         so jump right in!
    </p>
	</div>
</div>

<div class="row" style="padding-top: 75px"><div class='col-md-4 col-md-offset-4'><h2 class="text-center feature-header">Features</h2></div></div>
<hr class="feature">

<div class="row">
    {% include feature.html text="Maintain your DNS data as a high-level DS, with macros, and variables for easier updates." img="biology.svg" %}
	{% include feature.html text="Super extensible! Plug-in architecture makes adding new DNS providers and Registrars easy!" img="light-bulb.svg" %}
	{% include feature.html text="Eliminate vendor lock-in. Switch DNS providers easily, any time, with full fidelity." img="group.svg" %}
	{% include feature.html text="Reduce points of failure: Easily maintain dual DNS providers and easily drop one that is down." img="layers.svg" %}
	{% include feature.html text="Supports 10+ DNS Providers including BIND, AWS Route 53, Google DNS, and name.com" img="cancel.svg" %}
	{% include feature.html text="Apply CI/CD principles to DNS: Unit-tests, system-tests, automated deployment." img="share.svg" %}
	{% include feature.html text="All the benefits of Git (or any VCS) for your DNS zone data. View history. Accept PRs." img="document.svg" %}
	{% include feature.html text="Optimize DNS with SPF optimizer. Detect too many lookups. Flatten includes." img="mail.svg" %}
	{% include feature.html text="Runs on Linux, Windows, Mac, or any operating system supported by Go." img="speech-bubble.svg" %}
	{% include feature.html text="Enable/disable Cloudflare proxying (the \"orange cloud\" button) directly from your DNSControl files." img="cloud-computing.svg" %}
	{% include feature.html text="Assign an IP address to a constant and use the variable name throughout the configuration. Need to change the IP address globally? Just change the variable and \"recompile.\"" img="compass.svg" %}
	{% include feature.html text="Keep similar domains in sync with transforms, macros, and variables." img="attachment.svg" %}
</div>

<hr class="feature">

<div class="container-fluid">
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
				<li>
				    <a href="{{site.github.url}}/cli-variables">CLI variables</a>: Passing variables from CLI to JS
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
					<a href="{{site.github.url}}/js">Language Reference</a>: Full language description
				</li>
				<li>
					<a href="{{site.github.url}}/alias">Aliases</a>: ALIAS/ANAME records
				</li>
				<li>
					<a href="{{site.github.url}}/spf-optimizer">SPF Optimizer</a>: Optimize your SPF records
				</li>
				<li>
					<a href="{{site.github.url}}/caa-builder">CAA Builder</a>: Build CAA records the easy way
				</li>
				<li>
					<a href="{{site.github.url}}/get-certs">Let's Encrypt</a>: Renew your SSL/TLS certs
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
					<a href="{{site.github.url}}/why-the-dot">Why CNAME/MX/NS targets require a "trailing dot"</a>
				</li>
				<li>
					<a href="{{site.github.url}}/unittests">Testing</a>: Unit Testing for you DNS Data
				</li>
				<li>
					<a href="{{site.github.url}}/notifications">Notifications</a>: Be alerted when your domains are changed
				</li>
				<li>
					<a href="{{site.github.url}}/code-tricks">Code Tricks</a>: Safely use macros and loops.
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
				<li>
					<a href="{{site.github.url}}/release-engineering">Release Engineering</a>: How to build and ship a release
				</li>
				<li>
					<a href="{{site.github.url}}/bug-triage">Bug Triage</a>: How bugs are triaged
				</li>
			</ul>
		</div>
	</div>
</div>

<hr class="feature">

<p><small>Icons made by Freepik from <a href="http://www.flaticon.com">www.flaticon.com</a></small></p>
