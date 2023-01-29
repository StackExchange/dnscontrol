package route53

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	r53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func (r *route53Provider) findOriginal(nameFQDN string, kType string) (r53Types.ResourceRecordSet, error) {
	for _, rec := range r.originalRecords {
		if unescape(rec.Name) == nameFQDN {
			recType := string(rec.Type)
			if recType == kType || "R53_ALIAS_"+recType == kType {
				return rec, nil
			}
		}
	}
	return r53Types.ResourceRecordSet{}, fmt.Errorf("no record set found to delete. Name: '%s'. Type: '%s'", nameFQDN, kType)
}

// reorderInstructions returns changes reordered to comply with AWS's requirements:
//   - The R43_ALIAS updates must come after records they refer to.  To handle
//     this, we simply move all R53_ALIAS instructions to the end of the list, thus
//     guaranteeing they will happen after the records they refer to have been
//     reated.
func reorderInstructions(changes diff2.ChangeList) diff2.ChangeList {
	var main, tail diff2.ChangeList
	for _, change := range changes {
		if change.Key.Type == "R53_ALIAS" {
			tail = append(tail, change)
		} else {
			main = append(main, change)
		}
	}
	return append(main, tail...)
	// NB(tlim): This algorithm is O(n*2) but it is simple and usually only
	// operates on very small lists.
}
