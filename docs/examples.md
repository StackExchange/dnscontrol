---
layout: default
title: Examples
---

# Examples

* TOC
{:toc}

## Typical DNS Records

```js
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
```

## Set TTLs

```js
var mailTTL = TTL('1h');

D('example.com', registrar,
    NAMESERVER_TTL('10m'), // On domain apex NS RRs
    DefaultTTL('5m'), // Default for a domain

    MX('@', 5, '1.2.3.4', mailTTL), // use variable to
    MX('@', 10, '4.3.2.1', mailTTL), // set TTL

    A('@', '1.2.3.4', TTL('10m')), // individual record
    CNAME('mail', 'mx01') // TTL of 5m, as defined per DefaultTTL()
);
```

## Variables for common IP Addresses

```js
var addrA = IP('1.2.3.4')

D('example.com', REG, DnsProvider('R53'),
    A('@', addrA), // 1.2.3.4
    A('www', addrA + 1), // 1.2.3.5
)
```

NOTE: The `IP()` function doesn't currently support IPv6 (PRs welcome!).  IPv6 addresses are strings.

```js
var addrAAAA = "0:0:0:0:0:0:0:0";
```

## Variables to swap active Data Center

```js
var dcA = IP('5.5.5.5');
var dcB = IP('6.6.6.6');

// switch to dcB to failover
var activeDC = dcA;

D('example.com', REG, DnsProvider('R53'),
    A('@', activeDC + 5), // fixed address based on activeDC
)
```

## Macro to for repeated records

```js
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
```

## Add comments along complex SPF records

You normally can't put comments in the middle of a string,
but with a little bit of creativity you can document
each element of an SPF record this way.

```js
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
```

## Dual DNS Providers

```js
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
```

## Set default records modifiers

```js
DEFAULTS(
    NAMESERVER_TTL('24h'),
    DefaultTTL('12h'),
    CF_PROXY_DEFAULT_OFF
);
```
# Advanced Examples

## Automate Fastmail DKIM records

In this example we need a macro that can dynamically change for each domain.

Suppose you have many domains that use Fastmail as an MX. Here's a macro that sets the MX records.

```
var FASTMAIL_MX = [
  MX('@', 10, 'in1-smtp.messagingengine.com.'),
  MX('@', 20, 'in2-smtp.messagingengine.com.'),
]
```

Fastmail also supplied CNAMES to implement DKIM, and they all match a pattern
that includes the domain name. We can't use a simple macro. Instead, we use
a function that takes the domain name as a parameter to generate the right
records dynamically.

```
var FASTMAIL_DKIM = function(the_domain){
  return [
    CNAME('fm1._domainkey', 'fm1.' + the_domain + '.dkim.fmhosted.com.'),
    CNAME('fm2._domainkey', 'fm2.' + the_domain + '.dkim.fmhosted.com.'),
    CNAME('fm3._domainkey', 'fm3.' + the_domain + '.dkim.fmhosted.com.')
  ]
}
```

We can then use the macros as such:

```
D("example.com", REG_NONE, DnsProvider(DSP_R53_MAIN),
    FASTMAIL_MX,
    FASTMAIL_DKIM('example.com')
)
```
