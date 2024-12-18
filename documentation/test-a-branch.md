### Test A Branch

Instructions for testing DNSControl at a particular PR or branch.

## Using Docker

Using Docker assures you're using the latest version of Go and doesn't require you to install anything on your machine, other than Docker!

Assumptions:
* `/THE/PATH` -- Change this to the full path to where your dnsconfig.js and other files are located.
* `INSERT_BRANCH_HERE` -- The branch you want to test.  The branch associated with a PR is listed on [https://github.com/StackExchange/dnscontrol/branches](https://github.com/StackExchange/dnscontrol/branches).

```
docker run -it -v /THE/PATH:/dns golang
git clone https://github.com/StackExchange/dnscontrol.git
cd dnscontrol
git checkout INSERT_BRANCH_HERE
go install

cd /dns
dnscontrol preview
```

If you want to run the integration tests, follow the
[Integration Tests](https://docs.dnscontrol.org/developer-info/integration-tests) document
as usual. The directory to be in is `/go/dnscontrol/integrationTest`.

```
cd /go/dnscontrol/integrationTest
go test -v -verbose -provider INSERT_PROVIDER_NAME -start 1 -end 3
```

Change `INSERT_PROVIDER_NAME` to the name of your provider (BIND, ROUTE53, GCLOUD, etc.)


## Not using Docker

Assumptions:
* `/THE/PATH` -- Change this to the full path to where your dnsconfig.js and other files are located.
* `INSERT_BRANCH_HERE` -- The branch you want to test.  The branch associated with a PR is listed on [https://github.com/StackExchange/dnscontrol/branches](https://github.com/StackExchange/dnscontrol/branches).

Step 1: Install Go

[https://go.dev/dl/](https://go.dev/dl/)

Step 2: Check out the software

```
git clone https://github.com/StackExchange/dnscontrol.git
cd dnscontrol
git checkout INSERT_BRANCH_HERE
go install

cd /THE/PATH
dnscontrol preview
```

Step 3: Clean up

`go install` put the `dnscontrol` program in your `$HOME/bin` directory. You probably want to remove it.

```
$ rm -i $HOME/bin/dnscontrol
```

## Other useful docs

* How to run the integrations tests:
  * [https://docs.dnscontrol.org/developer-info/integration-tests](https://docs.dnscontrol.org/developer-info/integration-tests)
