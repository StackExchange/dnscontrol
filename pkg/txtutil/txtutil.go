package txtutil

import "github.com/StackExchange/dnscontrol/v3/models"

// SplitSingleLongTxt finds TXT records with a single long string and splits it
// into 255-octet chunks. This is used by providers that, when a user specifies
// one long TXT string, split it into smaller strings behind the scenes.
// Typically this replaces the TXTMulti capability.
func SplitSingleLongTxt(records []*models.RecordConfig) {
	for _, rc := range records {
		if rc.HasFormatIdenticalToTXT() {
			if len(rc.TxtStrings) == 1 {
				if len(rc.TxtStrings[0]) > 255 {
					rc.SetTargetTXTs(splitChunks(rc.TxtStrings[0], 255))
				}
			}
		}
	}
}

func splitChunks(buf string, lim int) []string {
	var chunk string
	chunks := make([]string, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}
