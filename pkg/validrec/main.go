package validrec

import "github.com/StackExchange/dnscontrol/v3/models"

type IsValid func(*models.RecordConfig) (bool, []error)

func FirstError(rc *models.RecordConfig, fns []IsValid) (bool, []error) {

	for i, fn := range fns {
		fn(rc)
	}
}

// Are fns returning "true" for error?  Is false, []error{thing} a
// warning?
