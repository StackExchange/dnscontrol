---
layout: default
title: Examples
---

# Examples

* TOC
{:toc}

## Typical DNS Records

{% highlight javascript %}

D('example.com', REG, DnsProvider('GCLOUD'),
    A('@', '1.2.3.4'),  // The naked or 'apex' domain.
    A('server1', '2.3.4.5'),
    AAAA('wide', '2001:0db8:85a3:0000:0000:8a2e:0370:7334'),
    CNAME('www', 'server1'),
    CNAME('another', 'service.mycloud.com.'),
    MX('mail', 10, 'mailserver'),
    MX('mail', 20, 'mailqueue'),
    TXT('the', 'message'),
    NS('department2', 'ns1.dnsexample.com.'), // use different nameservers
    NS('department2', 'ns2.dnsexample.com.') // for department2.example.com
)

{% endhighlight %}

## Set TTLs

{% highlight javascript %}

var mailTTL = TTL('1h');

D('example.com', registrar,
    NAMESERVER_TTL('10m'), // On domain apex NS RRs
    DefaultTTL('5m'), // Default for a domain

    MX('@', 5, '1.2.3.4', mailTTL), // use variable to
    MX('@', 10, '4.3.2.1', mailTTL), // set TTL

    A('@', '1.2.3.4', TTL('10m')), // individual record
    CNAME('mail', 'mx01') // TTL of 5m, as defined per DefaultTTL()
);

{% endhighlight %}

## Variables for common IP Addresses

{% highlight javascript %}

var addrA = IP('1.2.3.4')

D('example.com', REG, DnsProvider('R53'),
    A('@', addrA), // 1.2.3.4
    A('www', addrA + 1), // 1.2.3.5
)
{% endhighlight %}

## Variables to swap active Data Center

{% highlight javascript %}

var dcA = IP('5.5.5.5');
var dcB = IP('6.6.6.6');

// switch to dcB to failover
var activeDC = dcA;

D('example.com', REG, DnsProvider('R53'),
    A('@', activeDC + 5), // fixed address based on activeDC
)
{% endhighlight %}

## Macro to for repeated records

{% highlight javascript %}

var GOOGLE_APPS_RECORDS = [
    MX('@', 1, 'aspmx.l.google.com.'),
    MX('@', 5, 'alt1.aspmx.l.google.com.'),
    MX('@', 5, 'alt2.aspmx.l.google.com.'),
    MX('@', 10, 'alt3.aspmx.l.google.com.'),
    MX('@', 10, 'alt4.aspmx.l.google.com.'),
    CNAME('calendar', 'ghs.googlehosted.com.'),
    CNAME('drive', 'ghs.googlehosted.com.'),
    CNAME('mail', 'ghs.googlehosted.com.'),
    CNAME('groups', 'ghs.googlehosted.com.'),
    CNAME('sites', 'ghs.googlehosted.com.'),
    CNAME('start', 'ghs.googlehosted.com.'),
]

D('example.com', REG, DnsProvider('R53'),
   GOOGLE_APPS_RECORDS,
   A('@', '1.2.3.4')
)

{% endhighlight %}

## Add comments along complex SPF records

You normally can't put comments in the middle of a string,
but with a little bit of creativity you can document
each element of an SPF record this way.

{% highlight javascript %}

var SPF_RECORDS = TXT('@', [
    'v=spf1',
    'ip4:1.2.3.0/24',           // NY mail server
    'ip4:4.3.2.0/24',           // CO mail server
    'include:_spf.google.com',  // Google Apps
    'include:mailgun.org',      // Mailgun (requested by Ticket#12345)
    'include:servers.mcsv.net', // MailChimp (requested by Ticket#54321)
    'include:sendgrid.net',     // SendGrid (requested by Ticket#23456)
    'include:spf.mtasv.net',    // Desk.com (needed by IT team)
    '~all'
].join(' '));

D('example.com', REG, DnsProvider('R53'),
   SPF_RECORDS,
)

{% endhighlight %}

## Dual DNS Providers

{% highlight javascript %}

D('example.com', REG, DnsProvider('R53'), DnsProvider('GCLOUD'),
   A('@', '1.2.3.4')
)

// above zone uses 8 NS records total (4 from each provider dynamically gathered)
// below zone will only take 2 from each for a total of 4. May be better for performance reasons.

D('example2.com', REG, DnsProvider('R53',2), DnsProvider('GCLOUD',2),
   A('@', '1.2.3.4')
)

// or set a Provider as a non-authoritative backup (don't register its nameservers)
D('example3.com', REG, DnsProvider('R53'), DnsProvider('GCLOUD',0),
   A('@', '1.2.3.4')
)

{% endhighlight %}

## Set default records modifiers

{% highlight javascript %}

DEFAULTS(
	NAMESERVER_TTL('24h'),
	DefaultTTL('12h'),
	CF_PROXY_DEFAULT_OFF
);

{% endhighlight %}
