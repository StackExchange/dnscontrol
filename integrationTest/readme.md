### Integration Tests

This is a simple framework for testing dns providers by making real requests.

There is a sequence of changes that are defined in the test file that are run against your chosen provider.

For each step, it will run the config once and expect changes. It will run it again and expect no changes. This should give us much higher confidence that providers will work in real life.

## Configuration

`providers.json` should have an object for each provider type under test. This is identical to the json expected in creds.json for dnscontrol, except it also has a "domain" field specified for the domain to test. The domain does not even need to be registered for most providers. Note that `providers.json` expects environment variables to be specified with the relevant info.

## Running a test

1. Define all environment variables expected for the provider you wish to run. I setup a local `.env` file with the appropriate values and use [zoo](https://github.com/jsonmaur/zoo) to run my commands. 
2. run `go test -v -provider $NAME` where $NAME is the name of the provider you wish to run. 