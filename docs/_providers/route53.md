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
    "r53_main": {
        "KeyId": "your-aws-key",
        "SecretKey": "your-aws-secret-key",
        "Token": "optional-sts-token",
        "DelegationSet" : "optional-delegation-set-id"
    }
}
{% endhighlight %}

You can also use environment variables, but this is discouraged, unless your environment provides them already.

```
$ export AWS_ACCESS_KEY_ID=XXXXXXXXX
$ export AWS_SECRET_ACCESS_KEY=YYYYYYYYY
$ export AWS_SESSION_TOKEN=ZZZZZZZZ
```

{% highlight json %}
{
    "r53_main": {
        "KeyId": "$AWS_ACCESS_KEY_ID",
        "SecretKey": "$AWS_SECRET_ACCESS_KEY"
    }
}
{% endhighlight %}

Alternatively if you want to used [named profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html) you need to export the following variables

```
$ export AWS_SDK_LOAD_CONFIG=1
$ export AWS_PROFILE=ZZZZZZZZ
```

You can find some other ways to authenticate to Route53 in the [go sdk configuration](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html).

## Metadata
This provider does not recognize any special metadata fields unique to route 53.

## Usage
Example Javascript:

{% highlight js %}
var REG_NONE = NewRegistrar('none', 'NONE');
var R53 = NewDnsProvider('r53_main', 'ROUTE53');

D('example.tld', REG_NONE, DnsProvider(R53),
    A('test','1.2.3.4')
);
{% endhighlight %}

## Activation
DNSControl depends on a standard [AWS access key](https://aws.amazon.com/developers/access-keys/) with permission to list, create and update hosted zones. If you do not have the permissions required you will receive the following error message `Check your credentials, your not authorized to perform actions on Route 53 AWS Service`.

You can apply the `AmazonRoute53FullAccess` policy however this includes access to many other areas of AWS. The minimum permissions required are as follows:

{% highlight json %}
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "route53:CreateHostedZone",
                "route53:GetHostedZone",
                "route53:ListHostedZones",
                "route53:ChangeResourceRecordSets",
                "route53:ListResourceRecordSets",
                "route53:UpdateHostedZoneComment"
            ],
            "Resource": "*"
        }
    ]
}
{% endhighlight %}

If Route53 is also your registrar, you will need `route53domains:UpdateDomainNameservers` and `route53domains:GetDomainDetail` as well and possibly others.

## New domains
If a domain does not exist in your Route53 account, DNSControl will *not* automatically add it with the `push` command. You can do that either manually via the control panel, or via the command `dnscontrol create-domains` command.

## Delegation Sets
Creation of new delegation sets are not supported by this code. However, if you have a delegation set already created, ala:

```
$ aws route53 create-reusable-delegation-set --caller-reference "foo"
{
    "Location": "https://route53.amazonaws.com/2013-04-01/delegationset/12312312123",
    "DelegationSet": {
        "Id": "/delegationset/12312312123",
        "CallerReference": "foo",
        "NameServers": [
            "ns-1056.awsdns-04.org",
            "ns-215.awsdns-26.com",
            "ns-1686.awsdns-18.co.uk",
            "ns-970.awsdns-57.net"
        ]
    }
}
```

You can then reference the DelegationSet.Id in your `r53_main` block (with your other credentials) to have all created domains placed in that
delegation set.  Note that you you only want the portion of the `Id` after the `/delegationset/` (the `12312312123` in the example above).

> Delegation sets only apply during `create-domains` at the moment. Further work needs to be done to have them apply during `push`.

## Caveats

### Route53 errors if it is not the DnsProvider

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

If this happens to you, we'd appreciate it if you could help us fix the code. In the meanwhile, you can give the account additional IAM permissions so that it can do DNS-related actions, or simply use `NewRegistrar(..., 'NONE')` for now.

### Bug when converting new zones

You will see some weirdness if:

1.  A CNAME was created using the web UI
2.  The CNAME's target does NOT end with a dot.

What you will see: When dnscontrol tries to update such records, R53
only updates the first one.  For example if DNSControl is updating 3
such records, you will need to run `dnscontrol push` three times for
all three records to update.  Each time DNSControl is sending three
modify requests but only the first is executed.  After all such
records are modified by DNSControl, everything works as expected.

We believe this is a bug with R53.

This is only a problem for users converting old zones to DNSControl.

NOTE: When converting zones that include such records, the `get-zones`
command will generate `CNAME()` records without the trailing dot. You
should manually add the dot.  Run `dnscontrol preview` as normal to
check your work. However when you run `dnscontrol push` you'll find
you have to run it multiple times, each time one of those corrections
executes and the others do not.  Once all such records are replaced
this problem disappears.

More info is available in
[https://github.com/StackExchange/dnscontrol/issues/891](#891).


## Error messages

### Creds key mismatch

```
$ dnscontrol preview
Creating r53 dns provider: NoCredentialProviders: no valid providers in chain. Deprecated.
	For verbose messaging see aws.Config.CredentialsChainVerboseErrors
```

This means that the creds.json entry isn't found. Either there is no entry, or the entry name doesn't match the first parameter in the `NewDnsProvider()` call. In the above example, note
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

### Incomplete Signature

```
$  ./dnscontrol preview
IncompleteSignature: 'ABCDEFGHIJKLMNOPQRST/20200118/us-east-1/route53/aws4_request' not a valid key=value pair (missing equal-sign) in Authorization header: 'AWS4-HMAC-SHA256 Credential= ABCDEFGHIJKLMNOPQRST/20200118/us-east-1/route53/aws4_request, SignedHeaders=host;x-amz-date, Signature=571c0b13205669a338f0fb9f351dc03c7016c8737c738081bc885c68378ad877'.
        status code: 403, request id: 12a34b5c-d678-9e01-f2gh-3456i7jk89lm
```

This means a space is present in one or more of the credential values.
