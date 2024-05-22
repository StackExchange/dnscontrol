package hexonet

import (
	"fmt"

	"github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v4/response"
)

// GetHXApiError returns an error including API error code and error description.
func (n *HXClient) GetHXApiError(format string, objectid string, r *response.Response) error {
	return fmt.Errorf(format+" %q. [%v %s]", objectid, r.GetCode(), r.GetDescription())
}
