---
name: Google Cloud DNS
title: Google Cloud DNS Provider
layout: default
jsId: GCLOUD
---

# Google Cloud DNS Provider

## Configuration

To use this provider, add an entry to `creds.json` with `TYPE` set to `GCLOUD`.

For authentication you can either include a Service Account Key in the file or use Application Default Credentials (ADC)

### Using a Service Account Key
Copy the full JSON object into your `creds.json`. Newlines in the private key need to be replaced with `\n`.

Example:

```json
{
  "gcloud": {
    "TYPE": "GCLOUD",
    "type": "service_account",
    "project_id": "mydnsproject",
    "private_key_id": "a05483aa208364c56716b384efff33c0574d365b",
    "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADL2dhlY7YZbx7tpsfksOX\nih0DbxhiQ==\n-----END PRIVATE KEY-----\n",
    "client_email": "dnscontrolacct@mydnsproject.iam.gserviceaccount.com",
    "client_id": "107996619231234567750",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://accounts.google.com/o/oauth2/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dnscontrolsdfsdfsdf%40craigdnstest.iam.gserviceaccount.com",
    "name_server_set": "optional_name_server_set_name (contact your TAM)"
  }
}
```

**Note:** Don't confuse the `TYPE` and `type` fields.  `TYPE` is set to `GCLOUD` and specifies which provider type to use.  `type` specifies the type of account in use.

**Note**: The `project_id`, `private_key`, and `client_email`, are the only fields that are strictly required, but it is sometimes easier to just paste the entire JSON object in. Either way is fine.  `name_server_set` is optional and requires special permission from your TAM at Google in order to setup (See [Name server sets](#name_server_sets) below)

See [the Activation section](#activation) for some tips on obtaining these credentials.

### Using Application Default Credentials
If you prefer to authenticate using ADC you only need to specify `project_id` in your `creds.json` file.

Example:

```json
{
  "gcloud": {
    "TYPE": "GCLOUD",
    "project_id": "mydnsproject"
  }
}
```

**Note:** To use ADC, make sure to not add any `private_key` value to your configuration as that will prevent DNSControl from attempting to use ADC.

## Metadata
This provider does not recognize any special metadata fields unique to google cloud dns.

## Usage
An example `dnsconfig.js` configuration:

```js
var REG_NONE = NewRegistrar("name.com");
var DSP_GCLOUD = NewDnsProvider("gcloud");

D("example.tld", REG_NONE, DnsProvider(DSP_GCLOUD),
    A("test", "1.2.3.4")
);
```

## Activation
1. Go to your app-engine console and select the appropriate project.
2. Go to "API Manager > Credentials", and create a new "Service Account Key"

    <img src="{{ site.github.url }}/assets/gcloud-json-screen.png" alt="New Service Account" style="width: 900px;"/>

3. Choose an existing user, or create a new one. The user requires the "DNS Administrator" role.
4. Download the JSON key and copy it into your `creds.json` under the name of your gcloud provider.

## New domains
If a domain does not exist in your Google Cloud DNS account, DNSControl
will *not* automatically add it with the `push` command. You'll need to do that via the
control panel manually or via the `create-domains` command.

## Name server sets

This optional feature lets you pin domains to a set of GCLOUD name servers.  The `nameServerSet` field is exposed in their API but there is
currently no facility for creating a name server set.  You need special permission from your technical account manager at Google and they
will enable it on your account, responding with a list of names to use in the `name_server_set` field above.

> `name_server_set` only applies on `create-domains` at the moment. Additional work needs to be done to support it during `push`

# Debugging credentials

You can test your `creds.json` entry with the command: `dnscontrol check-creds foo GCLOUD` where `foo` is the name of key used in `creds.json`.  Error messages you might see:

* `googleapi: Error 403: Permission denied on resource project REDACTED., forbidden`
  * Hint: `project_id` may be invalid.
* `private key should be a PEM or plain PKCS1 or PKCS8; parse error:`
  * Hint: `private_key` may be invalid.
* `Response: {"error":"invalid_grant","error_description":"Invalid grant: account not found"}`
  * Hint: `client_email` may be invalid.
