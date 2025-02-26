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

Example:

```
var CNAME = recordBuilder('CNAME');
```

- Fix models/t_{type}.go

They should call RecordUpdateFields() (see t_mx.go for examples)

Remove unused functions:

```
fgrep --include='*.go' -r 'SetTargetDSStrings('
```



-- fix models/t_parse.go

Remove any mention of the type.

-- fix models/record.go

Remove any mention of the type in:
    func Downcase(recs []*RecordConfig) {
    func CanonicalizeTargets(recs []*RecordConfig, origin string) {


-- fix pkg/js/parse_tests

Any test that fails, copy `.json.ACTUAL` to `.json`
It should only (1) add Fields{},
remove .sub
update .targets to include the subdomain.

jstest.sh 014-caa.js


Once all types are using RawRecords...

Update pkg/normalize/validate.go
func validateRecordTypes(rec *models.RecordConfig, domain string, pTypes []string) error {

