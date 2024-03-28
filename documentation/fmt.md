# fmt

This is a stand-alone utility to pretty-format your `dnsconfig.js` configuration file.


```text
NAME:
   dnscontrol fmt - [BETA] Format and prettify a given file

USAGE:
   dnscontrol fmt [command options] [arguments...]

CATEGORY:
   utility

OPTIONS:
   --input value, -i value   Input file (default: "dnsconfig.js")
   --output value, -o value  Output file
   --help, -h                show help
```

## Examples

By default the output goes to stdout:

```
dnscontrol fmt >new-dnsconfig.js
```

You can also redirect the output via the `-o` option:

```
dnscontrol fmt -o new-dnsconfig.js
```

The **safest** method involves making a backup first:

```
cp dnsconfig.js dnsconfig.js.BACKUP
dnscontrol fmt -i dnsconfig.js.BACKUP -o dnsconfig.js
```

The **riskiest** method depends on the fact that DNSControl currently processes
the `-o` file after the input file is completely read.  It makes no backups.
This is useful if Git is your backup mechanism.

```
git commit -m'backup dnsconfig.js' dnsconfig.js
dnscontrol fmt -o dnsconfig.js
git diff -- dnsconfig.js
```
