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
        "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dnscontrolsdfsdfsdf%40craigdnstest.iam.gserviceaccount.com"
    }
}
{% endhighlight %}

**Note**: The `project_id`, `private_key`, and `client_email`, are the only fields that are strictly required, but it is sometimes easier to just paste the entire json object in. Either way is fine.

See [the Activation section](#activation) for some tips on obtaining these credentials.

## Metadata
This provider does not recognize any special metadata fields unique to google cloud dns.

## Usage
Use this provider like any other DNS Provider:

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var GCLOUD = NewDnsProvider("gcloud", GCLOUD);

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
will *not* automatically add it with the `create-domains` account. You'll need to do that via the
control panel manually.
