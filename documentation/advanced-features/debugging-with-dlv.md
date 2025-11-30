## Debugger

Test a particular function:

```shell
dlv test github.com/StackExchange/dnscontrol/v4/pkg/diff2 -- -test.run Test_analyzeByRecordSet
                                                ^^^^^^^^^
                                                Assumes you are in the pkg/diff2 directory.
```

Debug the integration tests:

```shell
dlv test github.com/StackExchange/dnscontrol/v4/integrationTest -- -test.v -test.run ^TestDNSProviders -verbose -profile BIND -start 7 -end 7
```

If you are using VSCode, the equivalent configuration is:

```
    "configurations": [
        {
            "name": "Debug Integration Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/integrationTest",
            "args": [
                "-test.v",
                "-test.run",
                "^TestDNSProviders",
                "-verbose",
                "-profile",
                "BIND",
                "-start",
                "7",
                "-end",
                "7"
            ],
            "buildFlags": "",
            "env": {},
            "showLog": true
        },
```
