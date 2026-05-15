## Configuration

{% hint style="info" %}
This provider is developed for the **Tencent Cloud API 3.0** platform.
{% endhint %}

This provider is for [Tencent Cloud DNS](https://cloud.tencent.com/product/dns) (DNSPod). To use this provider, add an entry to `creds.json` with `TYPE` set to `TENCENTDNS` along with your [API secrets](https://console.intl.cloud.tencent.com/cam/capi).

Example:

{% code title="creds.json" %}
```json
{
  "tencentdns": {
    "TYPE": "TENCENTDNS",
    "secret_id": "YOUR_SECRET_ID",
    "secret_key": "YOUR_SECRET_KEY",
    "site": "cn | intl"
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
    "region": "ap-guangzhou",
    "site": "intl"
  }
}
```
{% endcode %}

Optionally, you can specify a `site` (defaults to `"cn"`). Use `"intl"` for Tencent Cloud International accounts:

{% code title="creds.json" %}
```json
{
  "tencentdns": {
    "TYPE": "TENCENTDNS",
    "secret_id": "YOUR_SECRET_ID",
    "secret_key": "YOUR_SECRET_KEY",
    "site": "intl"
  }
}
```
{% endcode %}

Valid `site` values are:

- `cn`: Tencent Cloud mainland China APIs.
- `intl`: Tencent Cloud International APIs.

The `site` setting affects both DNSPod DNS management and registrar nameserver updates.

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
- **Tencent Cloud Site**: Use `site: "intl"` for Tencent Cloud International site, use `site: "cn"` for Tencent Cloud China site.
- **Line Management**: All records are created on the "默认" (Default) line.
- **New Domains**: DNSControl will automatically create non-existent domains in your account.
