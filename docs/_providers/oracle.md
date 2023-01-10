---
name: Oracle Cloud
title: Oracle Cloud Provider
layout: default
jsId: ORACLE
---
# Oracle Cloud Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `ORACLE`
along with other authentication parameters.

Create an API key through the Oracle Cloud portal, and provide the user OCID, tenancy OCID, key fingerprint, region, and the contents of the private key.
The OCID of the compartment DNS resources should be put in can also optionally be provided.

Example:

```json
{
  "oracle": {
    "TYPE": "ORACLE",
    "compartment": "$ORACLE_COMPARTMENT",
    "fingerprint": "$ORACLE_FINGERPRINT",
    "private_key": "$ORACLE_PRIVATE_KEY",
    "region": "$ORACLE_REGION",
    "tenancy_ocid": "$ORACLE_TENANCY_OCID",
    "user_ocid": "$ORACLE_USER_OCID"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Oracle Cloud.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("none");
var DSP_ORACLE = NewDnsProvider("oracle");

D("example.tld", REG_NONE, DnsProvider(DSP_ORACLE),
    NAMESERVER_TTL(86400),

    A("test", "1.2.3.4")
);
```

