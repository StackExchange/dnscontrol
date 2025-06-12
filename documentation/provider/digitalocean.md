## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `DIGITALOCEAN`
along with your [DigitalOcean Personal Access Token Token](https://cloud.digitalocean.com/account/api/tokens).

Example:

{% code title="creds.json" %}
```json
{
  "mydigitalocean": {
    "TYPE": "DIGITALOCEAN",
    "token": "your-digitalocean-token"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to DigitalOcean.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_DIGITALOCEAN = NewDnsProvider("mydigitalocean");

D("example.com", REG_NONE, DnsProvider(DSP_DIGITALOCEAN),
    A("test", "1.2.3.4"),
);
```
{% endcode %}

## Activation
[Create Personal Access Token](https://cloud.digitalocean.com/account/api/tokens)

Your access token must have access to create, read, update and delete domain records.

## Limitations

- Digitalocean DNS doesn't support `;` value with CAA-records ([DigitalOcean documentation](https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records/))
- While Digitalocean DNS supports TXT records with multiple strings,
  their length is limited by the max API request of 512 octets.
