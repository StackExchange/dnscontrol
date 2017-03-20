# convertzone -- Converts a standard DNS zonefile into tsv, pretty, or DSL

convertzone converts an old-style DNS zone file into one of three formats:

    convertzone [-mode=MODE] zonename [filename]

    -mode=tsv      Output the zone recoreds as tab-separated values
    -mode=pretty   Output the zone pretty-printed.
    -mode=dsl      Output the zone records as the DNSControl DSL language.

    zonename    The FQDN of the zone name.
    filename    File to read (optional. Defaults to stdin)

You must give the script both the zone name (i.e. "stackoverflow.com")
and the filename of the zonefile to read (or stdin). The domainname is not
automatically guessed from the input.

Output is sent to stdout.

Example:

    convertzone stackoverflow.com zone.stackoverflow.com >new/stackoverflow.com


## -mode=tsv:

This is useful for `awk` and other systems that expect a very
uniform set of input.

Example: Print all CNAMEs:

    convertzone -mode=tsv foo.com <zone.foo.com | awk '$4 == "CNAME" { print $1 " -> " $5 }'


## -mode=pretty:

This is useful for cleaning up a zonefile. It sorts the records,
moving SOA and NS records to the top of the zone; all other records
are alphabetically sorted; if a label has mutiple records, they are
listed in a logical (not numeric) order, multiple A records are
listed sorted by IP address, MX records are sorted by priority,
etc.  Use `-ttl` to set a default TTL.

Example: Clean up a zone file:

    convertzone -mode=pretty foo.com <old/zone.foo.com >new/zone.foo.com


## -mode=dsl:

This is useful for generating your draft `dnsconfig.js` configuration.
Pass the old zone through this program with `-mode=dsl` and append
it to your dnsconfig.js file. You'll probably need to clean it up
a bit: remove NS records (DnsProvider() inserts NS records for you,
change the order to be more logical and readable, manually check
over the results.

When converting a zonefile to DSL, we recommend first doing a
straightforward conversion, do not change any records at this time.
Now you can run `dnscontrol preview` to verify that dnsconfig.js
file is correct, and you will see that it has found zero changes
are needed. That means you have done the conversion property.  After
that step, do any cleanups you'd like to do (remove obsolete records,
etc.).  If you do such cleanups earlier in the process you can't
be entirely sure you've done the conversion correctly.

Example: Convert a zone filem and add it to your configuration:

    convertzone -mode=dsl foo.com <old/zone.foo.com >>dnsconfig.js
    #
    dnscontrol preview
    vim dnsconfig.js
    # (repeat until all warnings/errors are resolved)
    #
    # When everything is as you wish, push the changes live:
    dnscontrol push
