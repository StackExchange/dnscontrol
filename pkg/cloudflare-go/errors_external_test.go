package cloudflare_test

import (
	"testing"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/stretchr/testify/assert"
)

func TestError_CreateErrors(t *testing.T) {
	baseErr := &cloudflare.Error{
		StatusCode: 400,
		ErrorCodes: []int{10000},
	}

	requestErr := cloudflare.NewRequestError(baseErr)
	assert.True(t, requestErr.InternalErrorCodeIs(10000))
	limitError := cloudflare.NewRatelimitError(baseErr)
	assert.True(t, limitError.InternalErrorCodeIs(10000))
	svcErr := cloudflare.NewServiceError(baseErr)
	assert.True(t, svcErr.InternalErrorCodeIs(10000))
	authErr := cloudflare.NewAuthenticationError(baseErr)
	assert.True(t, authErr.InternalErrorCodeIs(10000))
	authzErr := cloudflare.NewAuthorizationError(baseErr)
	assert.True(t, authzErr.InternalErrorCodeIs(10000))
	notFoundErr := cloudflare.NewNotFoundError(baseErr)
	assert.True(t, notFoundErr.InternalErrorCodeIs(10000))
}
