## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `TRANSIP`
along with your TransIP credentials.

### Key Pairs

You can login with your `AccountName` and a `PrivateKey` which can be generated in the [TransIP control panel](https://www.transip.nl/cp/account/api/). The `PrivateKey` is a stringified version of the Private Key given by the API, see the example below, each newline is replaced by "\n".

Example:

{% code title="creds.json" %}
```json
{
  "transip": {
    "TYPE": "TRANSIP",
    "AccountName": "your-account-name",
    "PrivateKey": "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp\nwmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5\n1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh\n3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2\npIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX\nGukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il\nAkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF\nL0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k\nX6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl\nU9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ\n37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=\n-----END RSA PRIVATE KEY-----"
  }
}
```
{% endcode %}

### Access tokens

Or you can choose to have an `AccessToken` as credential. These can be generated in the [TransIP control panel](https://www.transip.nl/cp/account/api/) and have a limited lifetime

{% code title="creds.json" %}
```json
{
  "transip": {
    "TYPE": "TRANSIP",
    "AccessToken": "your-transip-personal-access-token"
  }
}
```
{% endcode %}


## Metadata

This provider does not recognize any special metadata fields unique to TransIP.

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_TRANSIP = NewDnsProvider("transip");

D("example.com", REG_NONE, DnsProvider(DSP_TRANSIP),
    A("test", "1.2.3.4")
);
```
{% endcode %}

## Activation

TransIP depends on a TransIP personal access token.

## Limitations

> "When multiple or none of the current DNS entries matches, the response will be an error with http status code 406." — _[TransIP - REST API - Update single DNS entry](https://api.transip.nl/rest/docs.html#domains-dns-patch)_

This makes it not possible, for example, to update a [`CAA()`](../language-reference/domain-modifiers/CAA.md) record in one update. Instead, the old DNS entry is deleted and the replacement is added. You'll see `[1/2]` and `[2/2]` in the DNSControl output whenever this happens.

### Example with a `CAA_BUILDER()`

{% code title="dnsconfig.js" %}
```diff
CAA_BUILDER({
    label: '@',
    iodef: 'mailto:info@cafferata.dev',
+   iodef_critical: true,
    issue: [
        'letsencrypt.org',
    ],
    issuewild: 'none',
}),
```
{% endcode %}

```shell
dnscontrol push --domains cafferata.dev
```

```shell
******************** Domain: cafferata.dev
2 corrections (transip)
#1: [1/2] delete: ± MODIFY cafferata.dev CAA (0 iodef "mailto:info@cafferata.dev" ttl=86400) -> (128 iodef "mailto:info@cafferata.dev" ttl=86400)
SUCCESS!
#2: [2/2] create: ± MODIFY cafferata.dev CAA (0 iodef "mailto:info@cafferata.dev" ttl=86400) -> (128 iodef "mailto:info@cafferata.dev" ttl=86400)
SUCCESS!
Done. 2 corrections.
```
