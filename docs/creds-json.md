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
    "apikey": "REDACTED",
    "apiuser": "REDACTED"
  },
  "inside": {
    "directory": "inzones",
    "filenameformat": "db_%T%?_%D"
  },
  "hexonet": {
    "apilogin": "$HEXONET_APILOGIN",
    "apipassword": "$HEXONET_APIPASSWORD",
    "debugmode": "$HEXONET_DEBUGMODE",
    "domain": "$HEXONET_DOMAIN"
  }
}
```

# Format

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

# Using a different name

The `--creds` flag allows you to specify a different file name.

* Normally the file is read as a JSON file.
* Do not end the filename with `.yaml` or `.yml` as some day we hope to support YAML.
* Rather than specifying a file, you can specify a program to be run. The output of the program must be valid JSON and will be read the same way.
  * If the name begins with `!`, the remainder of the name is taken to be the command to be run.
  * If the name is a file that is executable (chmod `+x` bit), it is taken as the command to be run.
  * Exceptions: The `x` bit is not checked if the filename ends with `.yaml`, `.yml` or `.json`.
  * Windows: Executing an external script isn't supported. There's no code that prevents it from trying, but it isn't supported.

# Don't store secrets in a Git repo!

Do NOT store secrets in a Git repository. That is not secure. For example,
storing the example `cloudflare_tal` is insecure because anyone with access to
your Git repository or the history will know your apiuser is `REDACTED`.
Removing secrets accidentally stored in Git is very difficult. You'll probably
give up and re-create the repo and lose all history.

Instead, use environment variables as in the `hexonet` example above.  Use
secure means to distribute the names and values of the environment variables.
