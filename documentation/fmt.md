# fmt

This is a stand-alone utility to pretty-format your `dnsconfig.js` configuration file.

```shell
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

```shell
dnscontrol fmt >new-dnsconfig.js
```

You can also redirect the output via the `-o` option:

```shell
dnscontrol fmt -o new-dnsconfig.js
```

The **safest** method involves making a backup first:

```shell
cp dnsconfig.js dnsconfig.js.BACKUP
dnscontrol fmt -i dnsconfig.js.BACKUP -o dnsconfig.js
```

The **riskiest** method depends on the fact that DNSControl currently processes
the `-o` file after the input file is completely read. It makes no backups.
This is useful if Git is your backup mechanism.

```shell
git commit -m'backup dnsconfig.js' dnsconfig.js
dnscontrol fmt -o dnsconfig.js
git diff -- dnsconfig.js
```
