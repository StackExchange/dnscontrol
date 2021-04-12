package hexonet

import (
	"fmt"

	"github.com/hexonet/go-sdk/response"
)

// GetHXApiError returns an error including API error code and error description.
func (n *HXClient) GetHXApiError(format string, objectid string, r *response.Response) error {
	return fmt.Errorf(format+" %s. [%s %s]", objectid, r.GetCode(), r.GetDescription())
}
