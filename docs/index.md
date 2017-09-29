---
layout: default
---

<div class="row jumbotron">
	<div class="col-md-12">
		<div>
			<h1 class="hometitle">DnsControl</h1>
			<p class="lead">DnsControl is a platform for seamlessly managing your DNS configuration across any number of DNS hosts, both in the cloud or in your own infrastructure. It manages all of the domains for the Stack Overflow network, and can do the same for you!</p>
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
         for more info.
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
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Maintain your DNS data as a high-level DS, with macros, and variables for easier updates.</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Super extensible! Plug-in architecture makes adding new DNS providers and Registrars easy!</p>
		</div>
	</div>
</div>

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Eliminate vendor lock-in. Switch between DNS providers easily.</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Reduce points of failure: Easily maintain  dual DNS providers.</p>
		</div>
	</div>
</div>

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Supports 10+ DNS Providers including BIND, AWS Route 53, Google DNS, and name.com</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Apply CI/CD principles to DNS: Unit-tests, system-tests, automated deployment.</p>
		</div>
	</div>
</div>

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">All the benefits of Git (or any VCS) for your DNS zone data.</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Optimize DNS with SPF optimizer (coming soon!)</p>
		</div>
	</div>
</div>

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Runs on Linux, Windows, Mac, or any operating system supported by Go.</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Enable/disable Cloudflare proxying (the "orange cloud" button) directly from your DNSControl files.</p>
		</div>
	</div>
</div>

<div class="row">
	<div class="col-md-6 left">
		<div class="col-md-2 left ">
			<img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;">
		</div>
		<div class="col-md-10">
			<p class="smaller">Assign an IP address to a constant and use the variable name throughout the configuration. Need to change the IP address globally? Just change the variable and "recompile."</p>
		</div>
	</div>
	<div class="col-md-6 right">
		<div class="col-md-2 left"><img class="fpicon" src="public/cog.svg" style="max-height: 40px; max-width: 40px;"></div>
		<div class="col-md-10">
		<p class="smaller">Keep similar domains in sync with transforms, macros, and variables.</p>
		</div>
	</div>
</div>


<div class="row" style="padding-top: 75px"><div class='col-md-4 col-md-offset-4'><h2 class="text-center feature-header"><a href="toc">Read More</a></h2></div></div>
<hr class="feature">
