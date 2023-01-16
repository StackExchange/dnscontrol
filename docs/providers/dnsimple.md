# DNSimple Provider
## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DNSIMPLE`
along with a DNSimple account access token.

You can also set the `baseurl` to use [DNSimple's free sandbox](https://developer.dnsimple.com/sandbox/) for testing.

Examples:

```json
{
  "dnsimple": {
    "TYPE": "DNSIMPLE",
    "token": "your-dnsimple-account-access-token"
  },
  "dnsimple_sandbox": {
    "TYPE": "DNSIMPLE",
    "baseurl": "https://api.sandbox.dnsimple.com",
    "token": "your-sandbox-account-access-token"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to DNSimple.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_DNSIMPLE = NewRegistrar("dnsimple");
var DSP_DNSIMPLE = NewDnsProvider("dnsimple");

D("example.tld", REG_DNSIMPLE, DnsProvider(DSP_DNSIMPLE),
    A("test", "1.2.3.4")
);
```

## Activation
DNSControl depends on a DNSimple account access token.

## Caveats

### CAA

As of July 2022, the DNSimple DNS does not accept spaces in CAA records. Putting spaces in the record will result in a 400 Validation Failed error.

```text
0 issue "letsencrypt.org; validationmethods=dns-01; accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

Removing the spaces will work.
```text
0 issue "letsencrypt.org;validationmethods=dns-01;accounturi=https://acme-v02.api.letsencrypt.org/acme/acct/1234"
```

