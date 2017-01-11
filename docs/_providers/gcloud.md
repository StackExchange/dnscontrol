---
name: "Google cloud DNS"
layout: default
jsId: GCLOUD
---

# Google cloud DNS Provider

## Configuration

In your providers config json file you must provide the following fields:
{% highlight json %}
{
 "gcloud":{
      "clientId": "abc123",
      "clientSecret": "abc123",
      "refreshToken":"abc123",
      "project": "your-gcloud-project-name",
 }
}
{% endhighlight %}

See [the Activation section](#activation) for some tips on obtaining these credentials.

## Metadata

This provider does not recognize any special metadata fields unique to googel cloud dns.

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

Because this provider depends on Oauth for authentication, generating the correct tokens can be a bit daunting. We recommend using the 
[Google Oauth2 Playground](https://developers.google.com/oauthplayground/) to generate refresh tokens.

1. In the google cloud platform console, create a project to host your DNS zones.
2. Go to API Manager / Credentials and create a new OAuth2 Client ID. Create it for a Web Application. 
  Make sure to add https://developers.google.com/oauthplayground to the "Authorized redirect URIs" section.

    ![New Oauth Client ID]({{ site.github.url }}/assets/gcloud-credentials.png)

3. Save your client id and client secret, along with your project name in your providers.json for DNSControl.
4. Go to the [Google Oauth2 Playground](https://developers.google.com/oauthplayground/). Click the settings icon on the top right side and select
"Use your own OAuth credentials". Enter your client id and client secret as obtained above.

    ![Settings Panel]({{ site.github.url }}/assets/gcloud-settings.png)

5. Select the scope for "Google Cloud DNS API v1 > https://www.googleapis.com/auth/ndev.clouddns.readwrite".
6. Make sure you authorize the api as the user you intend to make API requests with. 
7. Click "Exchange authorization code for tokens" and get a refresh and access token:

    ![Refresh Token]({{ site.github.url }}/assets/gcloud-token.png)
 
 8. Store the refresh token in your providers.json for DNSControl. It will take care of refreshing the token as needed.