package txtutil

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/decode"
)

// SplitSingleLongTxt finds TXT records with long strings and splits
// them into 255-octet chunks.
// This is used by providers that, when a user specifies
// one long TXT string, split it into smaller strings behind the scenes.
func SplitSingleLongTxt(records []*models.RecordConfig) {
	for _, rc := range records {
		if rc.HasFormatIdenticalToTXT() {
			rc.SetTargetTXTs(decode.Flatten255(rc.TxtStrings))
		}
	}
}
