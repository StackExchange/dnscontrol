package cloudflare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := map[string]struct {
		response []ResponseInfo
		want     string
	}{
		"basic complete response": {
			response: []ResponseInfo{{
				Code:    10000,
				Message: "Authentication error",
			}},
			want: "Authentication error (10000)",
		},
		"multiple complete response": {
			response: []ResponseInfo{
				{
					Code:    10000,
					Message: "Authentication error",
				},
				{
					Code:    10001,
					Message: "Not authentication error",
				},
			},
			want: "Authentication error (10000), Not authentication error (10001)",
		},
		"missing internal error code": {
			response: []ResponseInfo{{
				Message: "something is broke",
			}},
			want: "something is broke",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := &RequestError{cloudflareError: &Error{
				StatusCode: 400,
				Errors:     tc.response,
			}}

			assert.Equal(t, tc.want, got.Error())
		})
	}
}
