---
name: Name.com
title: Name.com Provider
layout: default
jsId: NAMEDOTCOM
---

# Name.com Provider

## Configuration
In your credentials file you must provide your name.com api username and access token:

{% highlight json %}
{
  "name.com":{
    "apikey": "yourApiKeyFromName.com",
    "apiuser": "yourUsername"
  }
}
{% endhighlight %}

There is another key name `apiurl` but it is optional and defaults to the correct value. If you want to use the test environment ("OT&E"), then add this:

    "apiurl": "https://api.dev.name.com",

export NAMEDOTCOM_URL='api.name.com'


## Metadata
This provider does not recognize any special metadata fields unique to name.com.

## Usage
**Example Javascript (DNS hosted with name.com):**

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var NAMECOM = NewDnsProvider("name.com","NAMEDOTCOM");

D("example.tld", REG_NAMECOM, DnsProvider(NAMECOM),
    A("test","1.2.3.4")
);
{%endhighlight%}


**Example Javascript (Registrar only. DNS hosted elsewhere):**

{% highlight js %}
var REG_NAMECOM = NewRegistrar("name.com","NAMEDOTCOM");
var R53 = NewDnsProvider("r53", "ROUTE53");

D("example.tld", REG_NAMECOM, DnsProvider(R53),
    A("test","1.2.3.4")
);
{%endhighlight%}

{% include alert.html text="Note: name.com does not allow control over the NS records of your zones via the api. It is not recommended to use name.com's dns provider unless it is your only dns host." %}

## Activation
In order to activate API functionality on your Name.com account, you must apply to the API program. The application form is [located here](https://www.name.com/reseller/apply). It usually takes a few days to get a response. After you are accepted, you should receive your API token via email.

## Tips and error messages

### invalid character '<'

```
integration_test.go:140: api returned unexpected response: invalid character '<' looking for beginning of value
```

This error means an invalid URL is being used to reach the API
endpoint.  It usually means a setting is `api.name.com/api` when
`api.name.com` is correct (i.e. remove the `/api`).

In integration tests:

 * Wrong: `export NAMEDOTCOM_URL='api.name.com/api'`
 * Right: `export NAMEDOTCOM_URL='api.name.com'`

In production, the `apiurl` setting in creds.json is wrong. You can
simply leave this option out and use the default, which is correct.

TODO(tlim): Improve the error message. (Volunteer needed!)


### dial tcp: lookup https: no such host

```
integration_test.go:81: Failed getting nameservers Get https://https//api.name.com/api/v4/domains/stackosphere.com?: dial tcp: lookup https: no such host
```

When running integration tests, this error
means you included the `https://` in the `NAMEDOTCOM_URL` variable.
You meant to do something like `export NAMEDOTCOM_URL='api.name.com' instead.

In production, the `apiurl` setting in creds.json needs to be
adjusted. You can simply leave this option out and use the default,
which is correct. If you are using the EO&T system, leave the
protocol (`http://`) off the URL.
