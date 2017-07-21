package normalize

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
)

func hasSpfRecords(recs []*models.RecordConfig) bool {
	for _, rec := range recs {
		if rec.Type == "TXT" && strings.HasPrefix(rec.Target, "v=spf1 ") {
			fmt.Println(rec)
			return true
		}
	}
	return false
}

func flattenSpf(domain *models.DomainConfig) error {
	fmt.Println("flattenSpf")
	return nil
}
