package hexonet

import (
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/hexonet/go-sdk/v3/response"
)

// GetHXApiError returns an error including API error code and error description.
func (n *HXClient) GetHXApiError(format string, objectid string, r *response.Response) error {
	return printer.Errorf(format+" %q. [%v %s]", objectid, r.GetCode(), r.GetDescription())
}
