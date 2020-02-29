---
name: Google Cloud DNS
title: Google Cloud DNS Provider 
layout: default
jsId: GCLOUD
---

# Google Cloud DNS Provider

## Configuration

For Google cloud authentication, DNSControl requires a JSON 'Service Account Key' for your project. Newlines in the private key need to be replaced with `\n`.Copy the full JSON object into your `creds.json` like so:

{% highlight json %}
{
    "gcloud": {
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
        "name_server_set" : "optional_name_server_set_name (contact your TAM)"
    }
}
{% endhighlight %}

**Note**: The `project_id`, `private_key`, and `client_email`, are the only fields that are strictly required, but it is sometimes easier to just paste the entire json object in. Either way is fine.  `name_server_set` is optional and requires special permission from your TAM at Google in order to setup (See [Name server sets](#name_server_sets) below)

See [the Activation section](#activation) for some tips on obtaining these credentials.

## Metadata
This provider does not recognize any special metadata fields unique to google cloud dns.

## Usage
Use this provider like any other DNS Provider:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var GCLOUD = NewDnsProvider("gcloud", "GCLOUD");

D("example.tld", REG_NAMECOM, DnsProvider(GCLOUD),
    A("test","1.2.3.4")
);
{%endhighlight%}

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
