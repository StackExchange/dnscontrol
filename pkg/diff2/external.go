package diff2

import "github.com/StackExchange/dnscontrol/v3/models"

func handsoff(existing, desired models.Records) models.Records {

	// 		for each any HANDS_OFF() records:
	// 			exts := query(existing, hands_off_args)
	// 			// Modifying a hands-off record is an error.
	// 			both := intersection(desired, exts)
	// 			if len(both) != 0: error.
	// 			// Add the query results to desired.
	// 			desired = append(desired, exts)

	return existing
}
