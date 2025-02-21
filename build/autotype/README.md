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

- edit  models/types.go

Edit `func Seal()` to include a case for CNAME

-- fix pkg/js/parse_tests

```
for i in parse_tests/0*.json ; do cp $i.ACTUAL $i ; done
```

Any test that fails, copy `.json.ACTUAL` to `.json`
It should only (1) add Fields{},
remove .sub
update .targets to include the subdomain.

-- fix models/target.go

Add to GetTargetField()
Add to SetTarget()
(add a SetTargetCNAME() function if needed)

Add to PopulateFromStringFunc
