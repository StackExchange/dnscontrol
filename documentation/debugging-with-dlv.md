## Debugger

Test a particular function:

```shell
dlv test github.com/StackExchange/dnscontrol/v3/pkg/diff2 -- -test.run Test_analyzeByRecordSet
                                                ^^^^^^^^^
                                                Assumes you are in the pkg/diff2 directory.
```

Debug the integration tests:

```shell
dlv test github.com/StackExchange/dnscontrol/v3/integrationTest -- -test.v -test.run ^TestDNSProviders -verbose -provider NAMEDOTCOM -start 1 -end 1 -diff2
```
