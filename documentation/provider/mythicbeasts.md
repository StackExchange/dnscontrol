This is the provider for [Mythic Beasts](https://www.mythic-beasts.com/) using its [Primary DNS API v2](https://www.mythic-beasts.com/support/api/dnsv2).

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `MYTHICBEASTS` along with a Mythic Beasts API key ID and secret.

Example:

{% code title="creds.json" %}
```json
{
  "mythicbeasts": {
    "TYPE": "MYTHICBEASTS",
	"keyID": "xxxxxxx",
	"secret": "xxxxxx"
  }
}
```
{% endcode %}

## Usage

For each domain:

* Domains must be added in the [web UI](https://www.mythic-beasts.com/customer/domains), and have DNS enabled.
* In Mythic Beasts' DNS management web UI, new domains will have set a default DNS template of "Mythic Beasts nameservers only". You must set this to "None".

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_MYTHIC = NewDnsProvider("mythicbeasts");

D("example.com", REG_NONE, DnsProvider(DSP_MYTHIC),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
