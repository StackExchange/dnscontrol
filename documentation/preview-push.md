# preview/push

`preview` reads the dnsconfig.js file (or equivalent), determines what changes are to be made, and
prints them.  `push` is the same but executes the changes.

```shell
NAME:
   dnscontrol preview - read live configuration and identify changes to be made, without applying them

USAGE:
   dnscontrol preview [command options] [arguments...]

CATEGORY:
   main

OPTIONS:
   --config value                                             File containing dns config in javascript DSL (default: "dnsconfig.js")
   --dev                                                      Use helpers.js from disk instead of embedded copy (default: false)
   --variable value, -v value [ --variable value, -v value ]  Add variable that is passed to JS
   --ir value                                                 Read IR (json) directly from this file. Do not process DSL at all
   --creds value                                              Provider credentials JSON file (or !program to execute program that outputs json) (default: "creds.json")
   --providers value                                          Providers to enable (comma separated list); default is all. Can exclude individual providers from default by adding '"_exclude_from_defaults": "true"' to the credentials file for a provider
   --domains value                                            Comma separated list of domain names to include
   --notify                                                   set to true to send notifications to configured destinations (default: false)
   --expect-no-changes                                        set to true for non-zero return code if there are changes (default: false)
   --no-populate                                              Use this flag to not auto-create non-existing zones at the provider (default: false)
   --full                                                     Add headings, providers names, notifications of no changes, etc (default: false)
   --bindserial value                                         Force BIND serial numbers to this value (for reproducibility) (default: 0)
   --report value                                             (push) Generate a JSON-formatted report of the number of changes made.
   --help, -h                                                 show help
```

* `--config name`
  * Specifies the name of the main configuration file, normally
`dnsconfig.js`.

* `--creds name`
  * Specifies the name of the credentials file, normally `creds.json`.
    Typically the file is read. If the executable bit is set, the file is
    executed and the output is used as the configuration. See
    [creds.json][creds-json.md] for details.

* `--providers name,name2`
  * Specifies a comma-separated list of providers to
    enable. The default is all providers. A provider can opt out of being in the
    default list by `"_exclude_from_defaults": "true"` to the credentials entry for
    that provider. In that case, the provider will only be activated if it is
    included in `--providers`.

* `--domains value`
  * Specifies a comma-separated list of domains to include.
    Typically all domains are included in `preview`/`push`. Wildcards are not
    permitted except `*` at the start of the entry. For example, `--domains
    example.com,*.in-addr.arpa` would include `example.com` plus all reverse lookup
    domains.

* `--v foo=bar`
  * Sets the variable `foo` to the value `bar` prior to
    interpreting the configuration file. Multiple `-v` options can be used.

* `--notify`
  * Enables sending notifications to the destinations configured in `creds.json`.

* `--dev`
  * Developer mode. Normally `helpers.js` is embedded in the dnscontrol
    executable. With this flag, the local file `helpers.js` is read instead.

* `--expect-no-changes`
  * If set, a non-zero exit code is returned if there are
    changes. Normally DNSControl sets the exit code based on whether or not there
    were protocol errors or other reasons the program can not continue. With this
    flag set, the exit code indicates if any changes were required. This is
    typically used with `preview` to allow scripts to determine if changes would
    happen if `push` was used. For example, one might want to run `dnscontrol
    preview --expect-no-changes` daily to determine if changes have been made to
    a domain outside of DNSControl.

* `--no-populate`
  * Do not auto-create non-existing zones at the provider.
    Normally non-existent zones are automatically created at a provider (unless the
    provider does not implement zone creation). This flag disables that feature.

* `--full`
  * Add headings, providers names, notifications of no changes, etc. to
    the output. Normally the output of `preview`/`push` is extremely brief. This
    makes the output more verbose. Useful for debugging.

* `--bindserial value`
  * Force BIND serial numbers to this value. Normally the
    BIND provider generates SOA serial numbers automatically. This flag forces the
    serial number generator to output the value specified for all domains. This is
    generally used for reproducibility in testing pipelines.

* `--report name`
  * (`push` only!)  Generate a machine-parseable report of
    performed corrections in the file named `name`. If no name is specified, no
    report is generated.

## ppreview/ppush

{% hint style="info" %}
Starting in v4.9
{% endhint %}

The `ppreview`/`ppush` subcommands are a preview of a future feature where zone
data is gathered concurrently. The commands will go away when
they replace the existing `preview`/`push` commands.

* `--cmode value`
  * Concurrency mode. Specifies what kind of providers should be run concurrently.
    * `default` -- Providers are run sequentially or concurrently depending on whether the provider is marked as having been tested to run concurrently.
    * `none` -- All providers are run sequentially. This is the safest mode. It can be used if a concurrency bug is discovered.
    * `all` -- This is unsafe. It runs all providers concurrently, even the ones that have not be validated to run concurrently. It is generally only used for demonstrating bugs.

## oldpreview/oldpush and release plan

The new concurrent implementation is 100% backwards compatible with the original
serial implementation.  To be cautious, however, both implementations will be
available in case bugs are discovered.

* The new (concurrent) code will be available by using the `ppreview`/`ppush`
  commands.  These commands will not be removed until a major release (v5.x or
  later).  The `p` stands for "parallel" (later we realized that "concurrent" is
  more accurate.)  
* The older (serial) code will be available by using the `oldpreview`/`oldpush`
  commands, to be introduced in v4.14.  These commands are a temporary work-around
  and will not be subject to SemVer.  They may be removed at any time.
* The `preview`/`push` commands will evolve from using the serial to the
  concurrent implementations.  

If you find a bug in the concurrent version:

1. [Report the bug ASAP!](https://github.com/StackExchange/dnscontrol/issues)
2. Add the `--mode none` flag to see if that works around the problem.
3. If that doesn't work, use `oldpreview`/`oldpush` until the bug is fixed.

Here is the release plan:

* Prior to v4.14:
  * oldpreview/oldpush: does not exist
  * preview/push: old (serial) implementation
  * **ppreview/ppush: new (concurrent) implementation**
* v4.14:
  * oldpreview/oldpush: old (serial) implementation (with warning it may go away)
  * preview/push: old (serial) implementation (with suggestion to test ppreview/ppush)
  * **ppreview/ppush: new (concurrent) implementation**
* v4.15:
  * oldpreview/oldpush: old (serial) implementation (with warning)
  * **preview/push: new (concurrent) implementation**
  * **ppreview/ppush: new (concurrent) implementation (with suggestion to use preview/push)**
* Future release (at least 2 months, target date Jan 1, 2024 or later)
  * oldpreview/oldpush: removed from code-base
  * **preview/push: new (concurrent) implementation**
  * **ppreview/ppush: new (concurrent) implementation (with suggestion to use preview/push)**