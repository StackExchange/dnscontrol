---
layout: default
---

# The DNS Control SPF Optimizer

SPF records are hints to email systems that help them determine if
an incoming email message might be spam.  The SPF records are placed
in DNS TXT records like so:

    $ dig +short google.com txt
    "v=spf1 include:_spf.google.com ~all"

SPF records are intentionally limited to 10 verbs that would cause
DNS lookups. In the above example the `include:_spf.google.com`
would cause a DNS lookup.  The reason for the "10 lookup limit" is
to make it difficult to leverage the SPF system to create a DDOS
attack on a DNS server.

At StackOverflow, we use many SaaS services and we reached the "10
lookup limit" years ago.  We would like to unroll or "inline" the
includes but it would become a maintenance nightmare. What if we
unrolled the SPF include required for Google GSuite and then Google
changed the contents of the SPF records?

We figured that DNSControl could do a better job.

# For the impatient

## Step 1: Define your SPF like this

    var SPF_LIST_NORMAL = [
        'v=spf1',
        'ip4:198.252.206.0/24',           // comment
        'ip4:192.111.0.0/24',             // comment
        'include:_spf.google.com',        // comment
        'include:mailgun.org',            // comment
        'include:spf-basic.fogcreek.com', // comment
        '~all'
    ].join(" ");
                                   // Change these to the ones that should be flattened:
    var SPF_NORMAL = [             //        VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV
        TXT("@", SPF_LIST_NORMAL, {flatten: "spf-basic.fogcreek.com,mailgun.org", split: "_spf%d"}),
        TXT("_rawspf", SPF_LIST_NORMAL) // keep unmodified availible for other tools
    ]

## Step 2: For a domain that needs that SPF record, include `SPF_NORMAL` as if it is a record.

    D('example.com', ...
      SPF_NORMAL,
      ...
    )

## Step 3: Push the changes

`dnscontrol preview` and `dnscontrol push` work as you'd expect.  However now
your SPF record will be optimized for you.

You might want to check out the web-based SPF tool described below.



## Better comments

Here's how we define our SPF record:

    var SPF_SO_LIST = [
        'v=spf1',
        'ip4:198.252.206.0/24',           // ny-mail*
        'ip4:192.111.0.0/24',             // co-mail*
        'include:_spf.google.com',        // GSuite
        'include:mailgun.org',            // Greenhouse.io
        'include:spf-basic.fogcreek.com', // Fogbugz
        'include:mail.zendesk.com',       // Zenddesk
        'include:servers.mcsv.net',       // MailChimp (Ticket#12345)
        'include:sendgrid.net',           // SendGrid
        'include:spf.mtasv.net',          // Desk.com
        '~all'
    ].join(" ");

    D('example.com', ...
      TXT("@", SPF_SO_LIST),
      ...
    )

The first thing you'll notice is that by defining it this way each
component can include a comment explaining what it is for.  This
is important because, and we're not kidding here, for a long time
we didn't know what `include:spf.mtasv.net` was for and we were
afraid to remove it.  Finally someone remembered that it was for
Desk.com and we breathed a sigh of relief.  You'll also notice that
the Mailchimp entry includes the ticket number of the request to
add it.  Now we can refer to that ticket to better understand the
history.

In summary, listing your SPF record like this makes it easier to
maintain a complex SPF record. Certainly you agree that this is
better than `var SPF_SO_LIST = 'v=spf1 'ip4:198.252.206.0/24 'ip4:192.111.0.0/24 'include:_spf.google.com 'include:mailgun.org 'include:spf-basic.fogcreek.com 'include:mail.zendesk.com 'include:servers.mcsv.net 'include:sendgrid.net include:spf.mtasv.net ~all'`

However, we can do better.

# Better macros

Because we don't want to have to remember the "@", and because we
use the same SPF record for multiple domains (any domain that is
attached to our GSuite account), we define a macro called SPF for
use with many domains:

    var SPF = [ TXT("@", SPF_SO_LIST) ]
    D('example.com', ...
      SPF,
      ...
    )
    D('otherexample.com', ...
      SPF,
      ...
    )

This is a lot less typing.  It is also less error-prone: you don't have to remember the `'@'`.

However, we can do better.

# SPF optimizer

As mentioned before, SPF records are intentionally limited to 10
verbs that would cause DNS lookups. This count includes recursive
includes. For example, if you use an `include:` that includes 5
other domains, that's 6 lookups.  That leaves you to only 4 more
lookups.

We figured that DNSControl could do better. It could analyze an SPF
record and flatten it to reduce the number of lookups.

However, we're very paranoid. If we break email, a lot of people
notice.  Therefore our "flattening" system has some safety rules:

* The system is "opt in". You must specify exactly which includes will be flattened. We recommend you only flatten the minimum needed.
* The flattening works off a cached copy of the DNS lookups. We are concerned
that if someone else's DNS server is down, the optimizer will break and you
won't be able to `dnscontrol push`, which would be very bad especially in
an emergency.  Therefore. the process runs off a file called FILLIN but will
warn you if the file needs updating.  The updates are easy to do (DNSControl generates
the new file for you to use).

So what does it look like?

Add metadata to the TXT records:

* `flatten: "foo,bar"`  (flatten include:foo and include:bar)
* `split: "_spf%d"`   (if additional DNS records must be generated, make the  labels `_spf1`, `_spf2`, `_spf3`, and so on.)

Here's an example:

    var SPF = [
        TXT("@", SPF_SO_LIST, {flatten: "spf-basic.fogcreek.com,spf.mtasv.net", split: "_spf%d"}),
        TXT("_rawspf", SPF_SO_LIST) // keep unmodified availible for other tools
    ]
    D('example.com', ...
      SPF,
      ...
    )

As a result:

* TXT record on `example.com` will be optimized.
* TXT record on `_rawspf.example.com` is the unoptimized version, used purely for demonstration purposes.

You'll notice that we only flatten 2 of all the includes. These are sufficient to get
us to only 10 lookups. They're also the 2 domains that SPF records are the least important.
Thus, if their SPF records change and we don't notice, we won't be too greatly affected.

# Operational Guide


FILL IN THE SEQUENCE OF COMMANDS TO MAINTAIN THE CACHE.


# Interactive mode

To help you decide what to flatten, load `docs/flattener/index.html`
into your web browser and you will be able to play with your SPF
records.  We suggest you flatten only the minimum required to reach
10 or fewer lookups.

This tool runs entirely in your browser.

Start interactive mode: [interactive SPF tool](flattener/index.html)

# Future

We'd like to add other optimizations such as:

* De-dup
* Remove overlapping CIDR blocks
