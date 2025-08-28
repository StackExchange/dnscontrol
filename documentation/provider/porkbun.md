## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `PORKBUN`
along with your `api_key` and `secret_key`. More info about authentication can be found in [Getting started with the Porkbun API](https://kb.porkbun.com/article/190-getting-started-with-the-porkbun-api).

Example:

{% code title="creds.json" %}
```json
{
  "porkbun": {
    "TYPE": "PORKBUN",
    "api_key": "your-porkbun-api-key",
    "secret_key": "your-porkbun-secret-key"
  }
}
```
{% endcode %}

Porkbun has quite strict API limits. If you experience errors with this provider (common when you have many domains), you can set one or both of `max_attempts` and `max_duration` in the credentials configuration.

Example:

{% code title="creds.json" %}
```json
{
  "porkbun": {
    "TYPE": "PORKBUN",
    "api_key": "your-porkbun-api-key",
    "secret_key": "your-porkbun-secret-key",
    "max_attempts": "10",
    "max_duration": "5m"
  }
}
```
{% endcode %}

The default for `max_attempts` is 5. There is no maximum duration by default, instead the provider will perform exponential backoff between 1 and 10 seconds, until `max_attempts` is reached. To retry indefinitely until `max_duration` is reached, set `max_attempts` to any value below 1.

## Metadata

This provider does not recognize any special metadata fields unique to Porkbun.

## Usage

An example configuration: (DNS hosted with Porkbun):

{% code title="dnsconfig.js" %}
```javascript
var REG_PORKBUN = NewRegistrar("porkbun");
var DSP_PORKBUN = NewDnsProvider("porkbun");

D("example.com", REG_PORKBUN, DnsProvider(DSP_PORKBUN),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

An example configuration: (Registrar only. DNS hosted elsewhere)

{% code title="dnsconfig.js" %}
```javascript
var REG_PORKBUN = NewRegistrar("porkbun");
var DSP_R53 = NewDnsProvider("r53");

D("example.com", REG_PORKBUN, DnsProvider(DSP_R53),
    A("test", "1.2.3.4"),
);
```
{% endcode %}
