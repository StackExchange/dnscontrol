package rtypes

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/rtypes/cfsingleredirect"
)

func PostProcess(domains []*models.DomainConfig) error {

	for _, dc := range domains {
		fmt.Printf("DOMAIN: %d %s\n", len(dc.Records), dc.Name)
		for _, rc := range dc.Records {
			if rc.Type == "rtype" {
				switch rc.Args[0] {
				case "CF_SINGLE_REDIRECT":
					err := cfsingleredirect.FromArgs(rc, rc.Args[1:])
					if err != nil {
						return err
					}
				default:
				}
			}
		}
	}

	return nil
}
