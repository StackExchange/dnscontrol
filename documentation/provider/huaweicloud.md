## Configuration


This provider is for the [Huawei Cloud DNS](https://www.huaweicloud.com/intl/en-us/product/dns.html)(Public DNS).  To use this provider, add an entry to `creds.json` with `TYPE` set to `HUAWEICLOUD`.
along with the API credentials.

Example:

{% code title="creds.json" %}
```json
{
  "huaweicloud": {
    "TYPE": "HUAWEICLOUD",
    "KeyId": "YOUR_ACCESS_KEY_ID",
    "SecretKey": "YOUR_SECRET_ACCESS_KEY",
    "Region": "YOUR_SERVICE_REGION"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Huawei Cloud DNS.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_HWCLOUD = NewDnsProvider("huaweicloud");

D("example.com", REG_NONE, DnsProvider(DSP_HWCLOUD),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
DNSControl depends on a standard [IAM User Access Key](https://support.huaweicloud.com/intl/en-us/usermanual-iam/iam_02_0003.html) with permission to list, create and update hosted zones.

You can apply the `DNS FullAccess` policy however this includes access to many other areas. The minimum permissions required are as follows:

```json
{
    "Version": "1.1",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dns:recordset:delete",
                "dns:recordset:create",
                "dns:zone:create",
                "dns:recordset:get",
                "dns:zone:list",
                "dns:recordset:update",
                "dns:recordset:list"
            ]
        }
    ]
}
```

If you don't know how to fill in the `Region` parameter, you can refer to the [endpoint page of huaweicloud](https://developer.huaweicloud.com/intl/en-us/endpoint?DNS). For reference, I'm using on its international site, the `Region` I need to fill in is `ap-southeast-1`.

If you're still not sure, you can log into Huaweicloud's website and open the [API Explorer](https://console-intl.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=ListPublicZones), find the `ListPublicZones` API, select a different Region and click Debug to try and find your Region!

## New domains
If a domain does not exist in your Huawei Cloud account, DNSControl will automatically add it with the `push` command. Don't worry about incurring fees, it's free!

## GeoDNS
Managing GeoDNS RRSet on huawei cloud (also called **Line** in Huawei Cloud DNS) is not supported in dnscontrol.
If your Zone needs to use GeoDNS, please create it manually in the console and use [IGNORE](../language-reference/domain-modifiers/IGNORE.md) modifiers in dnscontrol to prevent changing it.
