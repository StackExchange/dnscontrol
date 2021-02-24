package activedir

import "github.com/StackExchange/dnscontrol/v3/pkg/txt"

func (client *activedirProvider) IsValidTXT(rc *RecordConfig) (bool, []error) {
	return txt.OneNonNullShortString(rc)
}
