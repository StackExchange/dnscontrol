`GANDI_V5` uses the v5 API and can act as a registrar provider
or a DNS provider. It is only able to work with domains
migrated to the new LiveDNS API, which should be all domains.
API keys are assigned to particular users.  Go to User Settings,
"Manage the user account and security settings", the "Security"
tab, then regenerate the "Production API key".

* API Documentation: https://api.gandi.net/docs
* API Endpoint: https://api.gandi.net/

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `GANDI_V5`
along your Gandi.net API key. The [sharing_id](https://api.gandi.net/docs/reference/) is optional.

The `sharing_id` selects between different organizations which your account is
a member of; to manage domains in multiple organizations, you can use multiple
`creds.json` entries.

How to find the `sharing_id`: The sharing ID is the second hex string found in
the URL on the portal. Log into the Gandi website, click on "organizations" in
the leftnav, and click on the organization name.  The URL will be something
like:

```text
https://admin.gandi.net/organizations/[not this hex string]/PLTS/[sharing id]/profile
```

Example:

{% code title="creds.json" %}
```json
{
  "gandi": {
    "TYPE": "GANDI_V5",
    "apikey": "your-gandi-key",
    "sharing_id": "your-sharing_id"
  }
}
```
{% endcode %}

## Metadata
This provider does not recognize any special metadata fields unique to Gandi.

## Limitations
This provider does not support using `ALIAS` in combination with DNSSEC,
whether `AUTODNSSEC` or otherwise.

This provider only supports `ALIAS` on the `"@"` zone apex, not on any other
names.

## Usage
An example configuration:

{% code title="dnsconfig.js" %}
```javascript
var REG_GANDI = NewRegistrar("gandi");
var DSP_GANDI = NewDnsProvider("gandi");

D("example.com", REG_GANDI, DnsProvider(DSP_GANDI),
    A("test", "1.2.3.4"),
END);
```
{% endcode %}

If you are converting from the old "GANDI" provider,
simply change "GANDI" to "GANDI_V5" in `dnsconfig.js`.
Be sure to test with `dnscontrol preview` before running `dnscontrol push`.

## New domains
If a domain does not exist in your Gandi account, DNSControl will *not* automatically add it with the `create-domains` command. You'll need to do that via the web UI manually.


## Common errors

#### Error getting corrections

```text
Error getting corrections: 401: The server could not verify that you authorized to access the document you requested. Either you supplied the wrong credentials (e.g., bad api key), or your access token has expired
```

This is the error you'll see if your `apikey` in `creds.json` is wrong or invalid.

#### Domain does not exist in profile

```text
WARNING: Domain 'example.com' does not exist in the 'secname' profile and will be added automatically.
```

This error is caused by the internal `ListZones()` functions returning no domain names.  This is usually because your `creds.json` information is pointing at an empty organization or no organization.  The solution is to set
`sharing_id` in `creds.json`.

#### get-zones "nameonly" returns nothing

If a `dnscontrol get-zones --format=nameonly CredId - all` returns nothing,
this is usually because your `creds.json`  information is pointing at an empty
organization or no organization.  The solution is to set `sharing_id` in
`creds.json`.
