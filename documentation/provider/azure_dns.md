## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AZURE_DNS`
along with the API credentials.

Example:

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "AZURE_RESOURCE_GROUP",
    "TenantID": "AZURE_TENANT_ID",
    "ClientID": "AZURE_CLIENT_ID",
    "ClientSecret": "AZURE_CLIENT_SECRET"
  }
}
```
{% endcode %}

You can also use environment variables:

```shell
export AZURE_SUBSCRIPTION_ID=XXXXXXXXX
export AZURE_RESOURCE_GROUP=YYYYYYYYY
export AZURE_TENANT_ID=ZZZZZZZZ
export AZURE_CLIENT_ID=AAAAAAAAA
export AZURE_CLIENT_SECRET=BBBBBBBBB
```

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "$AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "$AZURE_RESOURCE_GROUP",
    "ClientID": "$AZURE_CLIENT_ID",
    "TenantID": "$AZURE_TENANT_ID",
    "ClientSecret": "$AZURE_CLIENT_SECRET"
  }
}
```
{% endcode %}

NOTE: The ResourceGroup is case sensitive.

## Metadata
This provider does not recognize any special metadata fields unique to Azure DNS.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_AZURE_MAIN = NewDnsProvider("azuredns_main");

D("example.com", REG_NONE, DnsProvider(DSP_AZURE_MAIN),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

## Activation
DNSControl depends on a standard [Client credentials Authentication](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest) with permission to list, create and update hosted zones.

## New domains
If a domain does not exist in your Azure account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.

## Caveats

The ResourceGroup is case sensitive.
