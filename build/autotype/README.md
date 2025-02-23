# How to submit PRs

```
gh pr create --base branch_gorec --title "Title" --body "Body"
```

# How to add a new rtype


- edit  build/autotype/hints.go

Add the hint for the type.

If we are converting a legacy type, list the legacy fields from RecordType.


- edit integrationTest/helpers_integration_test.go

Remove `func cname()`


- edit  pkg/js/helpers.js

Remove the definition of the type:

```
var CNAME = recordBuilder('CNAME');
```


- Fix models/t_{type}.go

They should call RecordUpdateFields() (see t_mx.go for examples)


-- fix models/t_parse.go

Add to PopulateFromStringFunc


--



-- fix pkg/js/parse_tests

Any test that fails, copy `.json.ACTUAL` to `.json`
It should only (1) add Fields{},
remove .sub
update .targets to include the subdomain.

jstest.sh 014-caa.js
