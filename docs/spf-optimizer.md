---
layout: default
title: SPF Optimizer
---

# SPF Optimizer

dnscontrol can optimize the SPF settings on a domain by flattening
(inlining) includes and removing duplicates.  dnscontrol also makes
it easier to document your SPF configuration.

**Warning:** Flattening SPF includes is risky.  Only flatten an SPF
setting if it is absolutely needed to bring the number of "lookups"
to be less than 10. In fact, it is debatable whether or not ISPs
enforce the "10 lookup rule".


## The old way

Here is an example of how SPF settings are normally done:

```
D("example.tld", REG, DNS, ...
  TXT("v=spf1 ip4:198.252.206.0/24 ip4:192.111.0.0/24 include:_spf.google.com include:mailgun.org include:spf-basic.fogcreek.com include:mail.zendesk.com include:servers.mcsv.net include:sendgrid.net include:450622.spf05.hubspotemail.net ~all")
)
```

This has a few problems:

* No comments. It is difficult to add a comment. In particular, we want to be able to list which ticket requested each item in the SPF setting so that history is retained.
* Ugly diffs.  If you add an element to the SPF setting, the diff will show the entire line changed, which is difficult to read.
* Too many lookups. The SPF RFC says that SPF settings should not require more than 10 DNS lookups. If we manually flatten (i.e. "inline") an include, we have to remember to check back to see if the settings have changed. Humans are not good at that kind of thing.

## The dnscontrol way

```
D("example.tld", REG, DSP, ...
  A("@", "10.2.2.2"),
  MX("@", "example.tld."),
  SPF_BUILDER({
    label: "@",
    overflow: "_spf%d",
    raw: "_rawspf",
    parts: [
      "v=spf1",
      "ip4:198.252.206.0/24", // ny-mail*
      "ip4:192.111.0.0/24", // co-mail*
      "include:_spf.google.com", // GSuite
      "include:mailgun.org", // Greenhouse.io
      "include:spf-basic.fogcreek.com", // Fogbugz
      "include:mail.zendesk.com", // Zenddesk
      "include:servers.mcsv.net", // MailChimp
      "include:sendgrid.net", // SendGrid
      "include:450622.spf05.hubspotemail.net", // Hubspot (Ticket# SREREQ-107)
      "~all"
    ],
    flatten: [
      "spf-basic.fogcreek.com", // Rational: Being deprecated. Low risk if it breaks.
      "450622.spf05.hubspotemail.net" // Rational: Unlikely to change without warning.
    ]
  }),
);
```

By using the `SPF_BUILDER` we gain many benefits:

* Comments can appear next to the element they refer to.
* Diffs will be shorter and more specific; therefore easier to read.
* Automatic flattening.  We can specify which includes should be flattened and dnscontrol will do the work. It will even warn us if the includes change.

## Syntax

When you want to specify SPF settings for a domain, use the
`SPF_BUILD()` function.

```
D("example.tld", REG, DSP, ...
  ...
  ...
  ...
  SPF_BUILDER({
    label: "@",
    overflow: "_spf%d",  // Delete this line if you don't want big strings split.
    raw: "_rawspf",  // Delete this line if the default is sufficient.
    parts: [
      "v=spf1",
      // fill in your SPF items here
      "~all"
    ],
    flatten: [
      // fill in any domains to inline.
    ]
  }),
  ...
  ...
);
```

The parameters are:

* `label:` The label of the first TXT record. (Optional. Default: `"@"`)
* `overflow:` If set, SPF strings longer than 255 chars will be split into multiple TXT records. The value of this setting determines the template for what the additional labels will be named. If not set, no splitting will occur and dnscontrol may generate TXT strings that are too long.
* `raw:` The label of the unaltered SPF settings. (Optional. Default: `"_rawspf"`)
* `parts:` The individual parts of the SPF settings.
* `flatten:` Which includes should be inlined. For safety purposes the flattening is done on an opt-in basis. If `"*"` is listed, all includes will be flattened... this might create more problems than is solves due to length limitations.

`SPR_BUILDER()` returns multiple `TXT()` records:

  * `TXT("@", "v=spf1 .... ~all")`
    *  This is the optimized configuration.
  * `TXT("_spf1", "...")`
    * If the optimizer needs to split a long string across multiple TXT records, the additional TXT records will have labels `_spf1`, `_spf2`, `_spf3`, etc.
  * `TXT("_rawspf", "v=spf1 .... ~all")`
    * This is the unaltered SPF configuration. This is purely for debugging purposes and is not used by any email or anti-spam system.  It is only generated if flattening is requested.


We recommend first using this without any flattening. Make sure
`dnscontrol preview` works as expected. Once that is done, add the
flattening required to reduce the number of lookups to 10 or less.

To count the number of lookups, you can use our interactive SPF
debugger at [https://stackexchange.github.io/dnscontrol/flattener/index.html](https://stackexchange.github.io/dnscontrol/flattener/index.html)


## Notes about the DNS Cache

dnscontrol keeps a cache of the DNS lookups performed during
optimization.  The cache is maintained so that the optimizer does
not produce different results depending on the ups and downs of
other people's DNS servers. This makes it possible to do `dnscontrol
push` even if your or third-party DNS servers are down.

The DNS cache is kept in a file called `spfcache.json`. If it needs
to be updated, the proper data will be written to a file called
`spfcache.updated.json` and instructions such as the ones below
will be output telling you exactly what to do:

```
$ dnscontrol preview
1 Validation errors:
WARNING: 2 spf record lookups are out of date with cache (_spf.google.com,_netblocks3.google.com).
Wrote changes to spfcache.updated.json. Please rename and commit:
    $ mv spfcache.updated.json spfcache.json
    $ git commit spfcache.json
```

In this case, you are being asked to replace `spfcache.json` with
the newly generated data in `spfcache.updated.json`.

Needing to do this kind of update is considered a validation error
and will block `dnscontrol push` from running.

Note: The instructions are hardcoded strings. The filenames will
not change.

Note: The instructions assume you use git. If you use something
else, please do the appropriate equivalent command.

## Caveats:

1. Dnscontrol 'gives up' if it sees SPF records it can't understand.
This includes: syntax errors, features that our spflib doesn't know
about, overly complex SPF settings, and anything else that we we
didn't feel like implementing.

2. The TXT record that is generated may exceed DNS limits.  dnscontrol
will not generate a single TXT record that exceeds DNS limits, but
it ignores the fact that there may be other TXT records on the same
label.  For example, suppose it generates a TXT record on the bare
domain (stackoverflow.com) that is 250 bytes long. That's fine and
doesn't require a continuation record.  However if there is another
TXT record (not an SPF record, perhaps a TXT record used to verify
domain ownership), the total packet size of all the TXT records
could exceed 512 bytes, and will require EDNS or a TCP request.

3. Dnscontrol does not warn if the number of lookups exceeds 10.
We hope to implement this some day.


## Advanced Technique: Interactive SPF Debugger

dnscontrol includes an experimental system for viewing
SPF settings:

[https://stackexchange.github.io/dnscontrol/flattener/index.html](https://stackexchange.github.io/dnscontrol/flattener/index.html)

You can also run this locally (it is self-contained) by opening
`dnscontrol/docs/flattener/index.html` in your browser.

You can use this to determine the minimal number of domains you
need to flatten to have fewer than 10 lookups.

The output is as follows:

1. The top part lists the domain as it current is configured, how
many lookups it requires, and includes a checkbox for each item
that could be flattened.

2. Fully flattened: This section shows the SPF configuration if you
fully flatten it. i.e. This is what it would look like if all the
checkboxes were checked. Note that this result is likely to be
longer than 255 bytes, the limit for a single TXT string.

3. Fully flattened split: This takes the "fully flattened" result
and splits it into multiple DNS records.  To continue to the next
record an include is added.


## Advanced Technique: Define once, use many

In some situations we define an SPF setting once and want to re-use
it on many domains. Here's how to do this:

```
var SPF_MYSETTINGS = SPF_BUILDER({
  label: "@",
  overflow: "_spf%d",
  raw: "_rawspf",
  parts: [
    "v=spf1",
    ...
    "~all"
  ],
  flatten: [
    ...
  ]
});

D("example.tld", REG, DSP, ...
    SPF_MYSETTINGS
);

D("example2.tld", REG, DSP, ...
     SPF_MYSETTINGS
);
```
