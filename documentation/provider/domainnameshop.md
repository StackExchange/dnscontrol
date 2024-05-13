## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DOMAINNAMESHOP`
along with your [Domainnameshop Token and Secret](https://www.domeneshop.no/admin?view=api).

Example:

{% code title="creds.json" %}
```json
{
  "mydomainnameshop": {
    "TYPE": "DOMAINNAMESHOP",
    "token": "your-domainnameshop-token",
    "secret": "your-domainnameshop-secret"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Domainnameshop.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DOMAINNAMESHOP = NewDnsProvider("mydomainnameshop");

D("example.com", REG_NONE, DnsProvider(DSP_DOMAINNAMESHOP),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
[Create API Token and secret](https://www.domeneshop.no/admin?view=api)

## Limitations

- Domainnameshop DNS only supports TTLs which are a multiple of 60.
