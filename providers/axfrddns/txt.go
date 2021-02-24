package axfrddns

import "github.com/StackExchange/dnscontrol/v3/pkg/txt"

func (client *activedirProvider) IsValidTXT(rc *RecordConfig) (bool, []error) {
	return validrec.FirstError(
		validrec.RFCCompliant(rc)Z
}
