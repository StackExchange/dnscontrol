---
name: INWX
jsId: INWX
---

# INWX

INWX.de is a Berlin-based domain registrar.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `INWX`
along with your INWX username and password.

Example:

```json
{
  "inwx": {
    "TYPE": "INWX",
    "password": "yourPassword",
    "username": "yourUsername"
  }
}
```

### Two factor authentication

INWX supports two factor authentication via TOTP and does not allow TOTP codes to be reused. This means that you will only be able to log into your INWX account once every 30 seconds.
You will hit this limitation in the following two scenarios:

* You run DNSControl twice very quickly (to e.g. first use preview and then push). Waiting for 30 seconds to pass between these two invocations will work fine though.
* You use INWX as both the registrar and the DNS provider. In this case, DNSControl will try to login twice too quickly and the second login will fail because a TOTP code will be reused. The only way to support this configuration is to use a INWX account without two factor authentication.

If you cannot work around these two limitation it is possible to contact the INWX support to request a sub-account for API access only without two factor authentication.
See issue [issue 848](https://github.com/StackExchange/dnscontrol/issues/848#issuecomment-692288859) for details.

If two factor authentication has been enabled you will also need to provide a valid TOTP number.
This can also be done via an environment variable:

```json
{
  "inwx": {
    "TYPE": "INWX",
    "password": "yourPassword",
    "totp": "$INWX_TOTP",
    "username": "yourUsername"
  }
}
```

and then you can run

```bash
INWX_TOTP=12345 dnscontrol preview
```

It is also possible to directly provide the shared TOTP secret using the key "totp-key" in `creds.json`.
This secret is only shown once when two factor authentication is enabled and you'll have to make sure to write it down then.

**Important Notes**:
* Anyone with access to this `creds.json` file will have *full* access to your INWX account and will be able to transfer and/or delete your domains
* Storing the shared secret together with the password weakens two factor authentication because both factors are stored in a single place.

```json
{
  "inwx": {
    "TYPE": "INWX",
    "password": "yourPassword",
    "totp-key": "yourTOTPSharedSecret",
    "username": "yourUsername"
  }
}
```


### Sandbox
You can optionally also specify sandbox with a value of 1 to
redirect all requests to the sandbox API instead:

```json
{
  "inwx":{
    "username": "yourUsername",
    "password": "yourPassword",
    "sandbox": "1"
  }
}
```

If sandbox is omitted or set to any other value the production
API will be used.


## Metadata
This provider does not recognize any special metadata fields unique to
INWX.

## Usage
An example `dnsconfig.js` configuration file
for `example.tld` registered with INWX
and delegated to Cloudflare:

```js
var REG_INWX = NewRegistrar("inwx");
var DSP_CF = NewDnsProvider("cloudflare");

D("example.tld", REG_INWX, DnsProvider(DSP_CF),
    A("test", "1.2.3.4")
);
```
