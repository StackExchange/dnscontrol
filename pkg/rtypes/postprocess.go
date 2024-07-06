package rtypes

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/rtypes/cfsingleredirect"
)

func PostProcess(domains []*models.DomainConfig) error {

	var err error

	for _, dc := range domains {
		fmt.Printf("DOMAIN: %d %s\n", len(dc.Records), dc.Name)

		for _, rawRec := range dc.RawRecords {
			rec := &models.RecordConfig{}

			switch rawRec.Type {
			case "CF_SINGLE_REDIRECT":
				err = cfsingleredirect.FromRaw(rec, rawRec.Args)
			default:
				return fmt.Errorf("unknown rawrec type=%q", rawRec.Type)
			}
			if err != nil {
				return err
			}

		}
	}

	return nil
}
