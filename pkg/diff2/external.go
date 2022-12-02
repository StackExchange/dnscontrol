package diff2

import "github.com/StackExchange/dnscontrol/v3/models"

func handsoff(existing, desired models.Records) models.Records {
	return existing
}
