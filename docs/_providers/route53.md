---
name: Route 53
layout: default
jsId: ROUTE53
---
# Amazon Route 53 Provider

## Configuration
You can specify the API credentials in the credentials json file:

{% highlight json %}
{
 "r53_main":{
      "KeyId": "your-aws-key",
      "SecretKey": "your-aws-secret-key"
 }
}
{% endhighlight %}

You can also use environment variables, but this is discouraged, unless your environment provides them already.

```
$ export AWS_ACCESS_KEY_ID=XXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=YYYYYYYYY
```

You can find some other ways to authenticate to Route53 in the [go sdk configuration](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

## Metadata
This provider does not recognize any special metadata fields unique to route 53.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none','NONE');
var R53 = NewDnsProvider('r53_main', 'ROUTE53');

D('example.tld', REG_NONE, DnsProvider(R53),
    A('test','1.2.3.4')
);
{%endhighlight%}

## Activation
DNSControl depends on a standard [AWS access key](https://aws.amazon.com/developers/access-keys/) with permission to list, create and update hosted zones.

## New domains
If a domain does not exist in your Route53 account, DNSControl will *not* automatically add it with the `create-domains` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.

## Caveats
This code may not function properly if a domain has R53 as a Registrar
but not as a DnsProvider.  The situation is described in
[PR#155](https://github.com/StackExchange/dnscontrol/pull/155).

In this situation you will see a message like:

```
----- Registrar: r53_main
Error getting corrections: AccessDeniedException: User: arn:aws:iam::868399730840:user/dnscontrol is not authorized to perform: route53domains:GetDomainDetail
  status code: 400, request id: 48b534a1-7902-11e7-afa6-a3fffd2ce139
Done. 1 corrections.
```

If this happens to you, we'd appreciate it if you could help us fix the code.  In the meanwhile, you can give the account additional IAM permissions so that it can do DNS-related actions, or simply use `NewRegistrar(..., 'NONE')` for now.

## Error messages

### Creds key mismatch

```
$ dnscontrol preview
Creating r53 dns provider: NoCredentialProviders: no valid providers in chain. Deprecated.
	For verbose messaging see aws.Config.CredentialsChainVerboseErrors
```

This means that the creds.json entry isn't found. Either there is no entry, or the entry name doesn't match the first parameter in the `NewDnsProvider()` call.  In the above example, note
that the string `r53_main` is specified in `NewDnsProvider('r53_main', 'ROUTE53')` and that is the exact key used in the creds file above.

### Invalid KeyId

```
$ dnscontrol preview
Creating r53_main dns provider: InvalidClientTokenId: The security token included in the request is invalid.
	status code: 403, request id: 8c006a24-e7df-11e7-9162-01963394e1df
```

This means the KeyId is unknown to AWS.

### Invalid SecretKey

```
$ dnscontrol preview
Creating r53_main dns provider: SignatureDoesNotMatch: The request signature we calculated does not match the signature you provided. Check your AWS Secret Access Key and signing method. Consult the service documentation for details.
	status code: 403, request id: 9171d89a-e7df-11e7-8586-cbea3ea4e710
```

This means the SecretKey is incorrect. It may be a quoting issue.
