package hexonet

import (
	lr "github.com/hexonet/go-sdk/response/listresponse"
	"github.com/pkg/errors"
)

// GetHXApiError returns an error including API error code and error description.
func (n *HXClient) GetHXApiError(format string, objectid string, r *lr.ListResponse) error {
	return errors.Errorf(format+" %s. [%s %s]", objectid, r.Code(), r.Description())
}
