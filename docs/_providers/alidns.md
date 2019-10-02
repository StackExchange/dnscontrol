---
  name: Alibaba Cloud DNS
  layout: default
  jsId: ALI_DNS
---

# Alibaba Cloud DNS Provider

## Configuration

You can specify the API credentials in the credentials json file:

{% highlight json %}
{
  "alidns_main":{
    "KeyId": "your-key",
    "SecretKey": "your-secret-key",
    "Token": "optional-sts-token"
  }
}
{% endhighlight %}

## Metadata

Record level metadata available:
   * `alidns_line`: DNS resolution line (Geographically, ISP, search engine. A full list can be obtained [here in Chinese](https://help.aliyun.com/document_detail/29807.html))

**CAVEATS**: Avoid using `alidns_line` right now if you have different `alidns_line` associated with more than one record share the same (type, name, target).

## Usage

Example Javascript:

{% highlight js %} 
var REG_NONE = NewRegistrar('none','NONE');
var ALIDNS = NewDnsProvider('alidns_main', 'ALI_DNS');

D('example.tld', REG_NONE, DnsProvider(ALIDNS),
    A('test','1.2.3.4')
);
{% endhighlight %}

Alibaba Cloud DNS provides two custom redirect records (available only in China with ICP registration record). 
    * `REDIRECT_URL`
    * `FORWARD_URL`
   
## Activation

DNSControl depends on a standard [Access Key](https://www.alibabacloud.com/help/doc-detail/29009.htm) with permission to list, create and update hosted zones.

For security, you grant read/write access for specific zone:

{% highlight json %}
{
    "Version": "1",
    "Statement": [
        {
            "Action": "alidns:*",
            "Resource": "acs:alidns:*:*:domain/example.com",
            "Effect": "Allow"
        },
        {
            "Action": "alidns:Describe*",
            "Resource": "acs:alidns:*:*:*",
            "Effect": "Allow"
        }
    ]
}
{% endhighlight %}

## New domains

If a domain does not exist in your Alibaba Cloud account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.
