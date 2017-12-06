---
layout: default
title: SPF Optimizer
---

# SPF Optimizer

dnscontrol can optimize the SPF settings on a domain by flattening
(inlining) includes and removing duplicates.  dnscontrol also makes
it easier to document your SPF records.

**Warning:** Flattening SPF includes is risky.  Only flatten an SPF
setting if it is absolutely needed to bring the number of "lookups"
to be less than 10. In fact, it is debatable whether or not ISPs
enforce the "10 lookup rule".


# The old way

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

# The dnscontrol way:

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

# Ok, so how do I use it?

When you want to specify SPF settings for a domain, use the
`SPF_BUILD()` function.

```
D("example.tld", REG, DSP, ...
  ...
  ...
  ...
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
  ...
  ...
);
```

The parameters are:

* `label:` The label of the SPF record. (Optional. Default: `"@"`)
* `overflow:` If the optimizer needs to continue the SPF settings to additional DNS records, this determines the template for what the additional labels will be named. (Optional. Default: `"_spf%d"`)
* `raw:` The label of the unaltered SPF settings. (Optional. Default: `"_rawspf"`)
* `parts:` The individual parts of the SPF settings.
* `flatten:` Which includes should be inlined. For safety purposes the flattening is done on an opt-in basis.

`SPR_BUILDER()` returns multiple `TXT()` records:

  * `TXT("@", "v=spf1 .... ~all")`
    *  This is the optimized record.
  * `TXT("_spf1", "...")`
    * If the optimizer needs to add continuation records, they will be generated with labels `_spf1`, `_spf2`, `_spf3`, etc. The "overflow" 
  * `TXT("_rawspf", "v=spf1 .... ~all")`
    * This is the unaltered record. This is purely for debugging purposes and is not used by any email or anti-spam system.  It is only generated if flattening is requested.


# DNS Cache

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


# Advanced: Define once, use many

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
