This is the provider for [Namecheap](https://www.namecheap.com/).

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `NAMECHEAP`
along with your Namecheap API username and key:

Example:

{% code title="creds.json" %}
```json
{
  "namecheap": {
    "TYPE": "NAMECHEAP",
    "apikey": "yourApiKeyFromNameCheap",
    "apiuser": "yourUsername"
  }
}
```
{% endcode %}

You can optionally specify BaseURL to use a different endpoint - typically the
sandbox:

{% code title="creds.json" %}
```json
{
  "namecheapSandbox": {
    "TYPE": "NAMECHEAP",
    "apikey": "yourApiKeyFromNameCheap",
    "apiuser": "yourUsername",
    "BaseURL": "https://api.sandbox.namecheap.com/xml.response"
  }
}
```
{% endcode %}

if BaseURL is omitted, the production namecheap URL is assumed.


## Metadata
This provider does not recognize any special metadata fields unique to
Namecheap.

## Usage
An example `dnsconfig.js` configuration:

```javascript
var REG_NAMECHEAP = NewRegistrar("namecheap");
var DSP_BIND = NewDnsProvider("bind");

D("example.tld", REG_NAMECHEAP, DnsProvider(DSP_BIND),
    A("test", "1.2.3.4")
);
```

Namecheap provides custom redirect records URL, URL301, and FRAME.  These
records can be used like any other record:

```javascript
var REG_NAMECHEAP = NewRegistrar("namecheap");
var DSP_NAMECHEAP = NewDnsProvider("namecheap");

D("example.tld", REG_NAMECHEAP, DnsProvider(DSP_NAMECHEAP),
  URL("@", "http://example.com/"),
  URL("www", "http://example.com/"),
  URL301("backup", "http://backup.example.com/")
)
```

## Activation
In order to activate API functionality on your Namecheap account, you must
enable it for your account and wait for their review process. More information
on enabling API access is [located
here](https://www.namecheap.com/support/api/intro.aspx).
