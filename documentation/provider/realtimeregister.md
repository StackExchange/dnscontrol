[realtimeregister.com](https://realtimeregister.com) is a domain registrar based in the Netherlands.

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `REALTIMEREGISTER`
along with your API-key. Further configuration includes a flag indicating BASIC or PREMIUM DNS-service and a flag
indicating the use of the sandbox environment

**Example:**

{% code title="creds.json" %}
```json
{
  "realtimeregister": {
    "TYPE": "REALTIMEREGISTER",
    "apikey": "abcdefghijklmnopqrstuvwxyz1234567890",
    "sandbox" : "0",
    "premium" : "0"
  }
}
```
{% endcode %}

If sandbox is omitted or set to any other value than "1" the production API will be used.
If premium is set to "1", you will only be able to update zones using Premium DNS. If it is omitted or set to any other value, you
will only be able to update zones using Basic DNS.

**Important Notes**:
* It is recommended to create a 'DNSControl' user in your account settings with limited permissions
(i.e. VIEW_DNS_ZONE, CREATE_DNS_ZONE, UPDATE_DNS_ZONE, VIEW_DOMAIN, UPDATE_DOMAIN), otherwise anyone with
access to this `creds.json` file might have *full* access to your RTR account and will be able to transfer or delete your domains.

## Metadata
This provider does not recognize any special metadata fields unique to Realtime Register.

## Usage
An example `dnsconfig.js` configuration file

{% code title="dnsconfig.js" %}
```javascript
var REG_RTR = NewRegistrar("realtimeregister");
var DSP_RTR = NewDnsProvider("realtimeregister");

D("example.com", REG_RTR, DnsProvider(DSP_RTR),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}
