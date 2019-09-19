---
  name: Azure DNS
  layout: default
  jsId: AZURE_DNS
---

# Azure DNS Provider

You can specify the API credentials in the credentials json file:

{% highlight json %}
{
 "azuredns_main":{
      "SubscriptionID": "AZURE_SUBSCRIPTION_ID",
      "ResourceGroup": "AZURE_RESOURCE_GROUP",
      "TenantID": "AZURE_TENANT_ID",
      "ClientID": "AZURE_CLIENT_ID",
      "ClientSecret": "AZURE_CLIENT_SECRET",
 }
}
{% endhighlight %}

You can also use environment variables, but this is discouraged, unless your environment provides them already.

```
$ export AZURE_SUBSCRIPTION_ID=XXXXXXXXX
$ export AZURE_RESOURCE_GROUP=YYYYYYYYY
$ export AZURE_TENANT_ID=ZZZZZZZZ
$ export AZURE_CLIENT_ID=AAAAAAAAA
$ export AZURE_CLIENT_SECRET=BBBBBBBBB
```

{% highlight json %}
{
 "azuredns_main":{
      "SubscriptionID": "$AZURE_SUBSCRIPTION_ID",
      "ResourceGroup": "$AZURE_RESOURCE_GROUP",
      "TenantID": "$AZURE_TENANT_ID",
      "ClientID": "$AZURE_CLIENT_ID",
      "ClientSecret": "$AZURE_CLIENT_SECRET",
 }
}
{% endhighlight %}

## Metadata
This provider does not recognize any special metadata fields unique to Azure DNS.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none','NONE');
var ADNS = NewDnsProvider('azuredns_main', 'AZURE_DNS');

D('example.tld', REG_NONE, DnsProvider(ADNS),
    A('test','1.2.3.4')
);
{%endhighlight%}

## Activation
DNSControl depends on a standard [Client credentials Authentication](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest) with permission to list, create and update hosted zones.

## New domains
If a domain does not exist in your Azure account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.


