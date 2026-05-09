## Configuration

{% hint style="info" %}
This provider is developed for the **Tencent Cloud API 3.0** platform.
{% endhint %}

This provider is for [Tencent Cloud DNS](https://cloud.tencent.com/product/cns) (DNSPod). To use this provider, add an entry to `creds.json` with `TYPE` set to `TENCENTDNS` along with your API credentials.

Example:

{% code title="creds.json" %}
```json
{
  "tencentdns": {
    "TYPE": "TENCENTDNS",
    "secret_id": "YOUR_SECRET_ID",
    "secret_key": "YOUR_SECRET_KEY"
  }
}
```
{% endcode %}

Optionally, you can specify a `region` (defaults to `"ap-guangzhou"`):

{% code title="creds.json" %}
```json
{
  "tencentdns": {
    "TYPE": "TENCENTDNS",
    "secret_id": "YOUR_SECRET_ID",
    "secret_key": "YOUR_SECRET_KEY",
    "region": "ap-guangzhou"
  }
}
```
{% endcode %}

## Usage

An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_TENCENT = NewRegistrar("tencentdns", "TENCENTDNS");
var DSP_TENCENT = NewDnsProvider("tencentdns", "TENCENTDNS");

D("example.com", REG_TENCENT, DnsProvider(DSP_TENCENT),
    A("@", "1.2.3.4"),
    CNAME("www", "example.com."),
    MX("@", 10, "mail.example.com."),
    TXT("test", "hello world")
);
```
{% endcode %}

## Important Notes

### Features

- **MX Records**: Priority and target are handled automatically.
- **Registrar Support**: Supports updating authoritative nameservers for domains registered with Tencent Cloud.
- **Line Management**: All records are created on the "默认" (Default) line.
- **New Domains**: DNSControl will automatically create non-existent domains in your account.
