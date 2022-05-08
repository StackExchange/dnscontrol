---
name: CSC Global
title: CSC Global Provider
layout: default
jsId: CSCGLOBAL
---
# CSC Global Provider

DNSControl's CSC Global provider supports being a Registrar. Support for being a DNS Provider is not included, although CSC Global's API does provide for this so it could be implemented in the future.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `CSCGLOBAL`.

In your `creds.json` file, you must provide your API key and user/client token. You can optionally provide an comma separated list of email addresses to have CSC Global send updates to.

Example:

```json
{
  "cscglobal": {
    "TYPE": "CSCGLOBAL",
    "api-key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "user-token": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "notification_emails": "test@exmaple.tld,hostmaster@example.tld"
  }
}
```

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_CSCGLOBAL = NewRegistrar("cscglobal");
var DSP_BIND = NewDnsProvider("bind");

D("example.tld", REG_CSCGLOBAL, DnsProvider(DSP_BIND),
  A("test", "1.2.3.4")
);
```

## Activation
To get access to the [CSC Global API](https://www.cscglobal.com/cscglobal/docs/dbs/domainmanager/api-v2/) contact your account manager.
