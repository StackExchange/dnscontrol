---
layout: default
title: creds.json file format
---

# creds.json

When dnscontrol interacts with a provider, any API keys, credentials, or other
configuration parameters required are stored in `creds.json`.   The file contains a set of key/value pairs for each configuration.  That is, since a provider can be used multiple times with different credentials, the file contains a section for each set of credentials.

Here's a sample file:

```json
{
  "cloudflare_tal": {
    "TYPE": "CLOUDFLAREAPI",
    "apikey": "REDACTED",
    "apiuser": "REDACTED"
  },
  "inside": {
    "TYPE": "BIND",
    "directory": "inzones",
    "filenameformat": "db_%T%?_%D"
  },
  "hexonet": {
    "TYPE": "HEXONET",
    "apilogin": "$HEXONET_APILOGIN",
    "apipassword": "$HEXONET_APIPASSWORD",
    "debugmode": "$HEXONET_DEBUGMODE",
    "domain": "$HEXONET_DOMAIN"
  }
}
```

## Format

* Primary keys: (e.g. `cloudflare_tal`, `inside`, `hexonet`)
  * ...refer to the first parameter in the `NewRegistrar()` or `NewDnsProvider()` functions in a dnsconfig.js file.
  * ...may include any printable character except colon (`:`)
  * Convention: all lower case, usually the name of the provider or the username at the provider or both.
* Subkeys: (e.g. `apikey`, `apiuser` and etc.)
  * ...are whatever the provider specifies.
  * ...can be credentials, secrets, or configuration settings. In the above examples the `inside` setting is configuration parameters for the BIND provider, not credentials.
  * A missing subkey is not an error. The value is the empty string.
* Values:
  * ...may include any JSON string value including the empty string.
  * If a subkey starts with `$`, it is taken as an env variable.  In the above example, `$HEXONET_APILOGIN` would be replaced by the value of the environment variable `HEXONET_APILOGIN` or the empty string if no such environment variable exists.

## New in v3.16:

The special subkey "TYPE" is used to indicate the provider type (NONE,
CLOUDFLAREAPI, GCLOUD, etc).

Prior to v3.16, the provider type is specified as the second argument to
`NewRegistrar()` and `NewDnsProvider()` in `dnsconfig.js` or as a command-line
argument in tools such as `dnscontrol get-zones`.

Starting in v3.16, specifying the provider type as `-` (in `NewRegistrar()`,
`NewDnsProvider()` or on the command line) instructs DNSControl to substitute
the `TYPE` value found in `creds.json` if it exists.

In 4.0, DNSControl will require the "TYPE" subkey in `creds.json` entry. Using
it as the second parameter to `NewRegistrar()`/`NewDnsProvider()` in
`dnsconfig.js` or on the command line will no longer be supported. This will
break backwards compatibility. Prior to 4.0, the various commands will output
warnings and suggestions to avoid compatibility issues during the transition.

## Error messages

### Missing

Message: `WARNING: For future compatibility, add this entry creds.json:...`

Message: `WARNING: For future compatibility, update the ... entry in creds.json by adding:...`

These messages indicates that this provider is not mentioned in `creds.json`.  In v4.0
all providers used in `dnsconfig.js` will require an entry in `creds.json`.

For a smooth transition, please update your `creds.json` file now.

Here is the minimal entry required:

```json
{
  "entryName": {
    "TYPE": "FILL_IN"
  }
}
```

### hyphen

Message: `ERROR: creds.json entry ... has invalid ... value ...`

This indicates the entry for `creds.json` has a TYPE value that is invalid.  It might be blank or a hyphen (`-`).  Change it to one of the all caps identifiers
in [the service provider list](https://stackexchange.github.io/dnscontrol/provider-list).

The fix is to correct the `TYPE` parameter in the `creds.json` entry.

### cleanup

Message: `INFO: In dnsconfig.js New*(..., ...) can be simplified to New*(...)`

This message indicates that the same provider name is specified in `dnsconfig.js` and `creds.json` and offers a suggestion for reducing the redundancy.

The fix is to update `dnsconfig.js` as suggested in the error.  Usually this is
to simply remove the second parameter to the function.

### mismatch

Message: `ERROR: Mismatch found! creds.json entry ... has ... set to ... but dnsconfig.js specifies New*(..., ...)`

This indicates that the provider type specifed in `creds.json` does not match the one specifed in `dnsconfig.js` or on the command line.

The fix is to change one to match the other.

### fixcreds

Message: `ERROR: creds.json entry ... is missing ...: ...`

This indicates that the creds.json file is missing.  The token `-` was used in
`dnsconfig.js` or on the command-line to indicate that the provider type is to
be found in `creds.json`. However no `TYPE` subkey was found in that entry in
`creds.json`.

The fix is to add a `TYPE` subkey to the `creds.json` entry.

### hyphen

Message: `ERROR: creds.json entry ... has invalid ... value ...`

This indicates that the type `-` was specified in `creds.json`, which is not
allowed. The token `-` indicates that the value is to be found in `creds.json`
thus it can not be used in`creds.json`.

The fix is to change the `TYPE` subkey entry in `creds.json` from `-` to
a valid service provider identifier, as listed
in [the service provider list](https://stackexchange.github.io/dnscontrol/provider-list).


## Using a different file name

The `--creds` flag allows you to specify a different file name.

* Normally the file is read as a JSON file.
* Do not end the filename with `.yaml` or `.yml` as some day we hope to support YAML.
* Rather than specifying a file, you can specify a program to be run. The output of the program must be valid JSON and will be read the same way.
  * If the name begins with `!`, the remainder of the name is taken to be the command to be run.
  * If the name is a file that is executable (chmod `+x` bit), it is taken as the command to be run.
  * Exceptions: The `x` bit is not checked if the filename ends with `.yaml`, `.yml` or `.json`.
  * Windows: Executing an external script isn't supported. There's no code that prevents it from trying, but it isn't supported.

## Don't store secrets in a Git repo!

Do NOT store secrets in a Git repository. That is not secure. For example,
storing the example `cloudflare_tal` is insecure because anyone with access to
your Git repository or the history will know your apiuser is `REDACTED`.
Removing secrets accidentally stored in Git is very difficult. You'll probably
give up and re-create the repo and lose all history.

Instead, use environment variables as in the `hexonet` example above.  Use
secure means to distribute the names and values of the environment variables.
