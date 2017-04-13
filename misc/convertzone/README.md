# convertzone -- Converts a standard DNS zonefile into tsv, pretty, or DSL

convertzone converts an old-style DNS zone file into one of three formats:

    convertzone [-mode=MODE] zonename [filename]

    -mode=tsv      Output the zone recoreds as tab-separated values
    -mode=pretty   Output the zone pretty-printed.
    -mode=dsl      Output the zone records as the DNSControl DSL language.

    zonename    The FQDN of the zone name.
    filename    File to read (optional. Defaults to stdin)

Output is sent to stdout.

The zonename is required as it can not be guessed automatically from the input.

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
The output can be appended to the `dnsconfig.js` file as a good first draft.
You'll probably need to clean it up
a bit:

* remove NS records. DnsProvider() inserts NS records for you.
* re-order the records to be more logical and readable. (remember that the last item in a list must not end with a comma)
* manually check over the results.

When converting a zonefile to DSL, we recommend a 2-phase process.  First
create a dnsconfig.js file that exactly replicates your existing configuration.
Only when that is complete should you make any changes to the DNS zone data.
This is not required, but it is a lot safer.

### Step 0: Build the software.

Build the software and install in your personal bin:

```cmd
$ cd misc/convertzone/
$ go build
$ cp convertzone ~/bin/.
```

### Step 1: Convert exactly as-is.

In this phase the goal is to create a dnsconfig.js that exactly
replicates the existing DNS data.

Edit dnsconfig.js until `dnscontrol preview` shows no errors and
no changes to be made. This means the conversion of your old DNS
data is correct.

Resist the temptation to clean up and old, obsolete, records or to
add anything new. Experience has shown that making changes at this
time leads to difficult-to-find errors.

If convertzone could have done a better job, please let us know!

### Step 2: Make any changes you desire.

Once `dnscontrol preview` lists no changes, do any cleanups
you want.  For example, remove obsolete records or add new ones.

### Example workflow

Example: Convert a zone file and add it to your configuration:

    # Note this command uses ">>" to append to dnsconfig.js.  Do
    # not use ">" as that will erase the existing file.
    convertzone -mode=dsl foo.com <old/zone.foo.com >>dnsconfig.js
    #
    dnscontrol preview
    vim dnsconfig.js
    # (repeat until all warnings/errors are resolved)
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
