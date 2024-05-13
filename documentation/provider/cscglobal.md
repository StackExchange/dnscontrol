DNSControl's CSC Global provider supports being a Registrar. Support for being a DNS Provider is not included, although CSC Global's API does provide for this so it could be implemented in the future.

{% hint style="info" %}
**NOTE**: Experimental support for being a DNS Provider is available.
However it is not recommended as updates take 5-7 minutes, and the
next update is not permitted until the previous update is complete.
Use it at your own risk.  Consider it experimental and undocumented.
{% endhint %}

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `CSCGLOBAL`.

In your `creds.json` file, you must provide your API key and user/client token. You can optionally provide an comma separated list of email addresses to have CSC Global send updates to.

Example:

{% code title="creds.json" %}
```json
{
  "cscglobal": {
    "TYPE": "CSCGLOBAL",
    "api-key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "user-token": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "notification_emails": "test@example.com,hostmaster@example.com"
  }
}
```
{% endcode %}

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_CSCGLOBAL = NewRegistrar("cscglobal");
var DSP_BIND = NewDnsProvider("bind");

D("example.com", REG_CSCGLOBAL, DnsProvider(DSP_BIND),
  A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
To get access to the [CSC Global API](https://www.cscglobal.com/cscglobal/docs/dbs/domainmanager/api-v2/) contact your account manager.
