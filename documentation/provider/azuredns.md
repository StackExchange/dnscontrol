## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `AZURE_DNS`, along with the necessary credentials. The provider supports three authentication methods:

1. **DefaultAzureCredential (Recommended)**: Simplifies authentication by leveraging Azure's credential chain (e.g., environment variables, managed identities, Azure CLI, etc.).
2. **Client ID and Secret**: Provides backward compatibility for users who prefer this method.
3. **OIDC (InteractiveBrowserCredential)**: Allows interactive login via the browser for specific scenarios.

### Example Configurations

#### **DefaultAzureCredential (Recommended)**

This method does not require explicit credentials in `creds.json` and leverages Azure's default authentication chain:
- Managed Identity (if running in Azure)
- Environment variables
- Azure CLI credentials

No additional setup is required in `creds.json`:

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "AZURE_RESOURCE_GROUP"
  }
}
```
{% endcode %}

You can also use environment variables:

```shell
export AZURE_SUBSCRIPTION_ID=XXXXXXXXX
export AZURE_RESOURCE_GROUP=YYYYYYYYY
```

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "$AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "$AZURE_RESOURCE_GROUP"
  }
}
```
{% endcode %}

#### **Client ID and Secret (Backward Compatibility)**

To use the client ID and secret-based authentication:

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

#### **OIDC (Interactive Browser Authentication)**

To enable OIDC for interactive login:

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "AZURE_RESOURCE_GROUP",
    "TenantID": "AZURE_TENANT_ID",
    "UseOIDC": "true"
  }
}
```
{% endcode %}

+You can also use environment variables:
```shell
export AZURE_SUBSCRIPTION_ID=XXXXXXXXX
export AZURE_RESOURCE_GROUP=YYYYYYYYY
export AZURE_TENANT_ID=ZZZZZZZZ
export UseOIDC=true
```

{% code title="creds.json" %}
```json
{
  "azuredns_main": {
    "TYPE": "AZURE_DNS",
    "SubscriptionID": "$AZURE_SUBSCRIPTION_ID",
    "ResourceGroup": "$AZURE_RESOURCE_GROUP",
    "TenantID": "$AZURE_TENANT_ID",
    "UseOIDC": "$UseOIDC"
  }
}
```
{% endcode %}


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
);
```
{% endcode %}

## Activation
DNSControl depends on a standard [Client credentials Authentication](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli?view=azure-cli-latest) with permission to list, create and update hosted zones.

## New domains
If a domain does not exist in your Azure account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.

## Caveats

The ResourceGroup is case sensitive.
