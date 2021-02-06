---
name: TXT
parameters:
  - name
  - contents
  - modifiers...
---

TXT adds an TXT record To a domain. The name should be the relative
label for the record. Use `@` for the domain apex.

The contents is either a single or multiple strings.  To
specify multiple strings, include them in an array.

TXT records with multiple strings are only supported by some
providers. DNSControl will produce a validation error if the
provider does not support multiple strings.

Each string is a JavaScript string (quoted using single or double
quotes).  The (somewhat complex) quoting rules of the DNS protocol
will be done for you.

Modifiers can be any number of [record modifiers](#record-modifiers) or json objects, which will be merged into the record's metadata.

{% include startExample.html %}
{% highlight js %}
    D("example.com", REGISTRAR, ....,
      TXT('@', '598611146-3338560'),
      TXT('listserve', 'google-site-verification=12345'),
      TXT('multiple', ['one', 'two', 'three']),  // Multiple strings
      TXT('quoted', 'any "quotes" and escapes? ugh; no worries!'),
      TXT('_domainkey', 't=y; o=-;'), // Escapes are done for you automatically.
      TXT('long', '#'.repeat(10), AUTOSPLIT) // Escapes are done for you automatically.
    );
{%endhighlight%}
{% include endExample.html %}


# Long and multiple strings

DNS RFCs limit TXT strings to 255 bytes, but you can have multiple
such strings.  Most applications blindly concatenate the strings but
some services that use TXT records join them with a space between each
substring (citation needed!).

Not all providers support multiple strings and those that do often put
limits on them.

Therefore, DNSControl requires you to explicitly mark TXT records that
should be split.

Here are some examples:

    VERY_LONG_STRING = 'Z'.repeat(300)

    // This will produce a validation-time error:
    TXT('long1', VERY_LONG_STRING),

    // String will be split on 255-byte boundaries:
    TXT('long', VERY_LONG_STRING, AUTOSPLIT),

    // String split manually:
    TXT('long', ['part1', 'part2', 'part3']),

NOTE: Old releases of DNSControl blindly sent long strings to
providers. Some gave an error at that time, others quietly truncated
the strings, and some silently split them into multiple short
strings.  If you see an error that mentions
`ERROR: txt target >255 bytes and AUTOSPLIT not set` this means you
need to add AUTOSPLIT to explicitly split the string manually.

An example error might look like this:

    2020/11/21 00:03:21 printIR.go:94: ERROR: txt target >255 bytes and AUTOSPLIT not set: label="20201._domainkey" index=0 len=424 string[:50]="v=DKIM1; k=rsa; t=s; s=email; p=MIIBIjANBgkqhkiG9w..."

