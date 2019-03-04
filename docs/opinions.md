---
layout: default
title: An Opinionated System
---

# DNSControl is an opinionated system

DNSControl is an opinionated system. That means that we have certain
opinions about how things should work.

This page documents those opinions.


# Opinion #1: DNS should be treated like code.

Code is written in a high-level language, version controlled,
commented, tested, and reviewed by a third party... and all of that
happens before it goes into production.

DNS information should be stored in a version control system, like
Git or Mercurial, and receive all the benefits of using VCS.  Changes
should be in the form of PRs that are approved by someone-other-than-you.

DNS information should be tested for syntax, pass unit tests and
policy tests, all in an automated CI system that assures all changes
are made the same way. (We don't provide a CI system, but DNSControl
makes it easy to use one; and not use one when an emergency update
is needed.)

Pushing the changes into production should be effortless, not
requiring people to know which domains are on which providers, or
that certain providers do things differently that others.  The
credentials for updates should be controlled such that anyone can
write a PR, but not everyone has access to the credentials.


# Opinion #2: Non-experts should be able to safely make DNS changes.

The goal of DNSControl is to create a system that is set up by DNS
experts like you, but updates and changes can be made by your
coworkers who aren't DNS experts.

Things your coworkers should not have to know:

Your coworkers should not have to know obscure DNS technical
knowledge.  That's your job.

Your coworkers should not have to know what happens in ambiguous
situations.  That's your job.

Your coworkers should be able to submit PRs to dnsconfig.js for you
to approve; preferably via a CI system that does rudimentary checks
before you even have to see the PR.

Your coworkers should be able to figure out the language without
much training. The system should block them from doing dangerous
things (even if they are technically legal).


# Opinion #3: dnsconfig.js are not zonefiles.

A zonefile can list any kind of DNS record. It has no judgement and
no morals. It will let you do bad practices as long as the bits are
RFC-compliant.

dnsconfig.js is a high-level description of your DNS zone data.
Being high-level permits the code to understand intent, and stop
bad behavior.

TODO: List an example.


# Opinion #4: All DNS is lowercase for languages that have such a concept.

DNSControl downcases all DNS names (domains, labels, and targets).  #sorrynotsorry

When the system reads dnsconfig.js or receives data from DNS providers,
the DNS names are downcased.

This reduces code complexity, reduces the number of edge-cases that must
be tested, and makes the system safer to operate.

Yes, we know that DNS is case insensitive.  See Opinion #3.


# Opinion #5: Users should state what they want, and DNSControl should do the rest.

When possible, dnsconfig.js lists a high-level description of what
is desired and the compiler does the hard work for you.

Some examples:

* Macros and iterators permit you to state something once, correctly, and repeat it many places.
* TXT strings are expressed as JavaScript strings, with no weird DNS-required special escape characters.  DNSControl does the escaping for you.
* Domain names with Unicode are listed as real Unicode.  Punycode translation is done for you.
* IP addresses are expressed as IP addresses; and reversing them to in-addr.arpa addresses is done for you.
* SPF records are stated in the most verbose way; DNSControl optimizes it for you in a safe, opt-in way.


# Opinion #6 If it is ambiguous in DNS, it is forbidden in DNSControl.

When there is ambiguity an expert knows what the system will do.
Your coworkers should not be expected to be experts. (See Opinion #2).

We would rather DNSControl error out than require users to be DNS experts.

For example:

We know that "bar.com." is a FQDN because it ends with a dot.

Is "bar.com" a FQDN? Well, obviously it is, because it already ends
with ".com" and we all know that "bar.com.bar.com" is probably not
what the user intended.

We know that "bar" is *not* an FQDN because it doesn't contain any dots.

Is "meta.xyz" a FQDN?

That's ambiguous.  If the user knows that "xyz" is a top level domain (TLD)
then it is obvious that it is a FQDN.  However, can anyone really memorize
all the TLDSs?  There used to be just gov/edu/com/mil/org/net and everyone
could memorize them easily.  As of 2000, there are many, many, more.
You can't memorize them all.  In fact, even before 2000 you couldn't
memorize them all. (In fact, you didn't even realize that we left out "int"!)

"xyz" became a TLD in June 2014.  Thus, after 2014 a system like DNSControl
would have to act differently.  We don't want to be surprised by changes
like that.

Therefore, we require all CNAME, MX, and NS targets to be FQDNs (they must
end with a "."), or to be a shortname (no dots at all).  Everything
else is ambiguous and therefore an error.

# Opinion #7 Hostnames don't have underscores

DNSControl prints warnings if a hostname includes an underscore
(`_`) because underscores are not permitted in hostnames.  

We want to prevent a naive user from including an underscore
when they meant to use a hyphen (`-`).

Hostnames are more restrictive than general DNS labels.
To quote [the Wikipedia entry on hostnames](https://en.wikipedia.org/wiki/Hostname#Restrictions_on_valid_hostnames)
"While a hostname may not contain other characters, such as the
underscore character (`_`), other DNS names may contain the
underscore. Systems such as DomainKeys and service records use
the underscore as a means to assure that their special character
is not confused with hostnames. For example,
`_http._sctp.www.example.com` specifies a service pointer for an
SCTP capable webserver host (www) in the domain example.com."

However that leads to an interesting problem. When is a DNS label
a hostname and when it it just a DNS label?  There is no way to
know for sure because code can't guess intention.

Therefore we print a warning if a label has an underscore in it,
unless the rtype is SRV, TLSA, TXT, or if the name starts with
certain prefixes such as `_dmarc`.  We're always willing to
[add more exceptions](https://github.com/StackExchange/dnscontrol/pull/453/files).
