package route53

import (
	"fmt"

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
