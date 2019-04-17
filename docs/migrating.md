---
layout: default
title: Migrating zones to DNSControl
---

# Migrating zones to DNSControl

This document explains how to migrate (convert) DNS zones from
other systems to DNSControl's `dnsconfig.js` file.

This document assumes you have DNSControl set up and working on at
least one zone.  You should have a working `dnsconfig.js` file and
`creds.json` file as explained in the
[Getting Started]({{site.github.url}}/getting-started) doc.

## General advice

First, use the
[Getting Started]({{site.github.url}}/getting-started) doc
so that you have a working `dnsconfig.js` with at least one domain.

We recommend migrating one zone at a time. Start with a small,
non-critical, zone first to learn the process.  Convert larger,
more important, zones as you gain confidence.

Experience has taught us that the best way to migrate a zone is
to create an exact duplicate first. That is, convert the old DNS records
with no changes.  It is tempting to clean up the data as you do the migration...
removing that old CNAME that nobody uses any more, or adding an
A record you discovered was missing. Resist that temptation.  If you make any
changes it will be difficult to tell which changes were intentional
and which are mistakes. During the migration you will know you are done
when `dnscontrol preview` says there are no changes needed. At that
point it is safe to do any cleanups.

## Create the first draft

Create the first draft of the `D()` statement either manually or
automatically.

For a small domain you can probably create the `D()` statements by
hand, possibly with your text editor's search and replace functions.
However, where's the fun in that?

The `convertzone` tool can automate 90% of the conversion for you. It
reads a BIND-style zone file or an OctoDNS-style YAML file and outputs a `D()` statement
that is usually fairly complete. You may need to touch it up a bit.

The convertzone command is in the `cmd/convertzone` subdirectory.
Build instructions are
[here](https://github.com/StackExchange/dnscontrol/blob/master/cmd/convertzone/README.md).

If you do not use BIND already, most DNS providers will export your
existing zone data to a file called the BIND zone file format.

For example, suppose you owned the `foo.com` domain and the zone file
was in a file called `old/zone.foo.com`. This command will convert the file:

    convertzone -out=dsl foo.com <old/zone.foo.com >first-draft.js

If you are converting an OctoDNS file, add the flag `-in=octodns`:

    convertzone -in=octodns -out=dsl foo.com <config/foo.com.yaml >first-draft.js

Add the contents of `first-draft.js` to `dnsconfig.js`

Run `dnscontrol preview` and see if it finds any differences.
Edit dnsconfig.js until `dnscontrol preview` shows no errors and
no changes to be made. This means the conversion of your old DNS
data is correct.

convertzone makes a guess at what to do with NS records.
An NS record at the apex is turned into a NAMESERVER() call, the
rest are left as NS().  You probably want to check each of them for
correctness.

Resist the temptation to clean up and old, obsolete, records or to
add anything new. Experience has shown that making changes at this
time leads to unhappy surprises, and people will blame DNSControl.
Of course, once `dnscontrol preview` runs cleanly, you can do any
kind of cleanups you want.  In fact, they should be easier to do
now that you are using DNSControl!

If convertzone could have done a better job, please
[let us know](https://github.com/StackExchange/dnscontrol/issues)!

## Example workflow

Here is an example series of commands that would be used
to convert a zone. Lines that start with `#` are comments.

    # Note this command uses ">>" to append to dnsconfig.js.  Do
    # not use ">" as that will erase the existing file.
    convertzone -out=dsl foo.com <old/zone.foo.com >>dnsconfig.js
    #
    dnscontrol preview
    vim dnsconfig.js
    # (repeat these two commands until all warnings/errors are resolved)
    #
    # When everything is as you wish, push the changes live:
    dnscontrol push
    # (this should be a no-op)
    #
    # Make any changes you do desire:
    vim dnsconfig.js
    dnscontrol preview
    # (repeat until all warnings/errors are resolved)
    dnscontrol push