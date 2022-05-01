---
  name: Azure DNS
  layout: default
  jsId: AZURE_DNS
---

# Azure DNS Provider

Specify the API credentials in the cred.json file:

```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "ClientID": "AZURE_CLIENT_ID",
    "ClientSecret": "AZURE_CLIENT_SECRET",
    "ResourceGroup": "AZURE_RESOURCE_GROUP",
    "SubscriptionID": "AZURE_SUBSCRIPTION_ID",
    "TenantID": "AZURE_TENANT_ID"
  }
}
```

You can also use environment variables:

```bash
export AZURE_SUBSCRIPTION_ID=XXXXXXXXX
export AZURE_RESOURCE_GROUP=YYYYYYYYY
export AZURE_TENANT_ID=ZZZZZZZZ
export AZURE_CLIENT_ID=AAAAAAAAA
export AZURE_CLIENT_SECRET=BBBBBBBBB
```

```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "ClientID": "$AZURE_CLIENT_ID",
    "ClientSecret": "$AZURE_CLIENT_SECRET",
    "ResourceGroup": "$AZURE_RESOURCE_GROUP",
    "SubscriptionID": "$AZURE_SUBSCRIPTION_ID",
    "TenantID": "$AZURE_TENANT_ID"
  }
}
```

## Metadata
This provider does not recognize any special metadata fields unique to Azure DNS.

## Usage
Example Javascript:

```js
var REG_NONE = NewRegistrar("none");
var ADNS = NewDnsProvider("azuredns_main");

D("example.tld", REG_NONE, DnsProvider(ADNS),
    A("test", "1.2.3.4")
);
```

## Activation
DNSControl depends on a standard [Client credentials Authentication](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest) with permission to list, create and update hosted zones.

## New domains
If a domain does not exist in your Azure account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.


