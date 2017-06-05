# convertzone -- Converts a standard DNS zonefile into tsv, pretty, or DSL

## Building the software

Build the software and install in your personal bin:

```cmd
$ cd cmd/convertzone
$ go build
$ cp convertzone ~/bin/.
```


## Usage Overview

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


### -mode=tsv:

This is useful for `awk` and other systems that expect a very
uniform set of input.

Example: Print all CNAMEs:

    convertzone -mode=tsv foo.com <zone.foo.com | awk '$4 == "CNAME" { print $1 " -> " $5 }'


### -mode=pretty:

This is useful for cleaning up a zonefile. It sorts the records,
moving SOA and NS records to the top of the zone; all other records
are alphabetically sorted; if a label has mutiple records, they are
listed in a logical (not numeric) order, multiple A records are
listed sorted by IP address, MX records are sorted by priority,
etc.  Use `-ttl` to set a default TTL.

Example: Clean up a zone file:

    convertzone -mode=pretty foo.com <old/zone.foo.com >new/zone.foo.com


### -mode=dsl:

This is useful for generating your draft `dnsconfig.js` configuration.
The output can be appended to the `dnsconfig.js` file as a good first draft.

Example: Generate statements for a dnsconfig.js file:

    convertzone -mode=dsl foo.com <old/zone.foo.com >first-draft.js

Note: The conversion is not perfect. You'll need to manually clean
it up and insert it into `dnsconfig.js`.  More instructions in the
DNSControl [migration doc]({site.github.url}}/migration).
