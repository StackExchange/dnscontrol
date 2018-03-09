### Integration Tests

This is a simple framework for testing dns providers by making real requests.

There is a sequence of changes that are defined in the test file that are run against your chosen provider.

For each step, it will run the config once and expect changes. It will run it again and expect no changes. This should give us much higher confidence that providers will work in real life.

## Configuration

`providers.json` should have an object for each provider type under test. This is identical to the json expected in creds.json for dnscontrol, except it also has a "domain" field specified for the domain to test. The domain does not even need to be registered for most providers. Note that `providers.json` expects environment variables to be specified with the relevant info.

## Running a test

1. Define all environment variables expected for the provider you wish to run. I setup a local `.env` file with the appropriate values and use [zoo](https://github.com/jsonmaur/zoo) to run my commands. 
2. run `go test -v -provider $NAME` where $NAME is the name of the provider you wish to run. 

Example:

```
$ egrep R53 providers.json 
    "KeyId": "$R53_KEY_ID",
    "SecretKey": "$R53_KEY",
    "domain": "$R53_DOMAIN"
$ export R53_KEY_ID="redacted"
$ export R53_KEY="also redacted"
$ export R53_DOMAIN="testdomain.tld"
$ go test -v -verbose -provider ROUTE53
```

WARNING: The records in the test domain will be deleted.  Only use
a domain that is not used in production. Some providers have a way
to run tests on domains that aren't registered (often a test
environment or a side-effect of the company not being a registrar).
In other cases we use a domain we squat on, or we register a domain
called `dnscontrol-$provider.com` just for testing.

ProTip: If you run these tests frequently (and we hope you do), you
should create a script that you can `source` to set these
variables. Be careful not to check this script into Git since it
contains credentials.
