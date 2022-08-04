package recordaudit

import "github.com/StackExchange/dnscontrol/v3/models"

type Auditor struct {
	records   []models.Records
	checksFor map[string][]checker
}

type checker = func(models.RecordConfig) error

// Add registers a function to call on each record of a given type.
func (aud *Auditor) Add(rtype string, fn Checker) {
	if aud.checksFor == nil {
		aud.checksFor = map[string][]checker{}
	}
	aud.checksFor[rtype] = append(aud.checksFor[rtype], fn)
	// SPF records get any checkers that TXT records do.
	if rtype == "TXT" {
		aud.checksFor[rtype] = append(aud.checksFor["SPF"], fn)
	}
}

// Audit performs the audit. For each record it calls each function in
// the list of checks.
func (aud *Auditor) Audit() (errs []error) {
	// No checks? Exit early.
	if aud.checksFor == nil {
		return nil
	}

	// For each record, call the checks for that type, gather errors.
	for _, rc := range aud.records {
		for _, f := range aud.checksFor[rc.Type] {
			e := f(rc)
			if e != nil {
				errs = append(errs, e)
			}
		}
	}

	return errs
}
