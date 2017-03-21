---
layout: default
---

# Examples

* TOC
{:toc}

## Variables for common IP Addresses

{% highlight javascript %}

var addrA = IP("1.2.3.4")

D("example.com", REG, DnsProvider("R53"),
    A("@", addrA), //1.2.3.4
    A("www", addrA + 1), //1.2.3.5
)
{% endhighlight %}

## Variables to swap active Data Center

{% highlight javascript %}

var dcA = IP("5.5.5.5");
var dcB = IP("6.6.6.6");

// switch to dcB to failover
var activeDC = dcA;

D("example.com", REG, DnsProvider("R53"),
    A("@", activeDC + 5), // fixed address based on activeDC
)
{% endhighlight %}

## Macro to group repeated records

{% highlight javascript %}

var GOOGLE_APPS_DOMAIN_MX = [
    MX('@', 1, 'aspmx.l.google.com.'),
    MX('@', 5, 'alt1.aspmx.l.google.com.'),
    MX('@', 5, 'alt2.aspmx.l.google.com.'),
    MX('@', 10, 'alt3.aspmx.l.google.com.'),
    MX('@', 10, 'alt4.aspmx.l.google.com.'),
]

D("example.com", REG, DnsProvider("R53"),
   GOOGLE_APPS_DOMAIN_MX,
   A("@", "1.2.3.4")
)
{% endhighlight %}

## Dual DNS Providers

{% highlight javascript %}

D("example.com", REG, DnsProvider("R53"), DnsProvider("GCLOUD"),
   A("@", "1.2.3.4")
)

// above zone uses 8 NS records total (4 from each provider dynamically gathered)
// below zone will only take 2 from each for a total of 4. May be better for performance reasons.

D("example2.com", REG, DnsProvider("R53",2), DnsProvider("GCLOUD",2),
   A("@", "1.2.3.4")
)

// or set a Provider as a non-authoritative backup (don't register its nameservers)
D("example3.com", REG, DnsProvider("R53"), DnsProvider("GCLOUD",0),
   A("@", "1.2.3.4")
)

{% endhighlight %}