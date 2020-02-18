
!!! NOTE: This command has been replaced by the "dnscontrol get-zones"
!!! subcommand. It can do everything convertzone does and more, with
!!! fewer bugs.  This command will be removed from the distribution soon.

# convertzone -- Converts a standard DNS zonefile into tsv, pretty, or DSL

This is a crude hack we put together to read a couple common zonefile
formats and output them in a few different formats.  Current input
formats are BIND zonefiles and OctoDNS "config" YAML files.  Current
output formats as BIND zonefiles, tab separated records, or a draft
DNSControl dnsconfig.js file. For dnsconfig.js, it does about 90%
of the work, but should be manually verified.

The primary purpose of this program is to convert BIND-style
zonefiles to DNSControl dnsconfig.js files.  Nearly all DNS Service
providers include the ability to export records as a BIND-style zonefile.
This makes it easy to import DNS data from other systems into DNSControl.
Later OctoDNS input was added because we had the parser (as part of
the OctoDNS provider), so why not use it?

## Building the software

Build the software and install in your personal bin:

```cmd
$ cd cmd/convertzone
$ go build
$ cp convertzone ~/bin/.
```


## Usage Overview

convertzone: Read and write DNS zone files.

convertzone [-in=INPUT] [-out=OUTPUT] zonename [filename]

Input format:
-in=bind      BIND-style zonefiles (DEFAULT)
-in=octodns   OctoDNS YAML "config" files.

Output format:

-out=dsl      DNSControl DSL language (dnsconfig.js) (DEFAULT)
-out=tsv      TAB-separated values
-out=pretty   pretty-printed (BIND-style zonefiles)

zonename    The FQDN of the zone name.
filename    File to read (optional. Defaults to stdin)

The DSL output format is useful for creating the first
draft of your dnsconfig.js when importing zones from
other services.

The TSV format makes it easy to process a zonefile with
shell tools.

The PRETTY format is just a nice way to clean up a zonefile.

If no filename is specified, stdin is assumed.
Output is sent to stdout.

The zonename is required as it can not be guessed automatically from the input.

Example:

    convertzone stackoverflow.com zone.stackoverflow.com >new/draft.js


### -out=tsv:

This is useful for `awk` and other systems that expect a very
uniform set of input.

Example: Print all CNAMEs:

    convertzone -out=tsv foo.com <zone.foo.com | awk '$4 == "CNAME" { print $1 " -> " $5 }'


### -out=pretty:

This is useful for cleaning up a zonefile. It sorts the records,
moving SOA and NS records to the top of the zone; all other records
are alphabetically sorted; if a label has mutiple records, they are
listed in a logical (not numeric) order, multiple A records are
listed sorted by IP address, MX records are sorted by priority,
etc.  Use `-ttl` to set a default TTL.

Example: Clean up a zone file:

    convertzone -out=pretty foo.com <old/zone.foo.com >new/zone.foo.com


### -out=dsl:

This is useful for generating your draft `dnsconfig.js` configuration.
The output can be appended to the `dnsconfig.js` file as a good first draft.

Example: Generate statements for a dnsconfig.js file:

    convertzone -out=dsl foo.com <old/zone.foo.com >first-draft.js

Note: The conversion is not perfect. You'll need to manually clean
it up and insert it into `dnsconfig.js`.  More instructions in the
DNSControl [migration doc]({site.github.url}}/migration).
