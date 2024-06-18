package cloudflare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceProperties(t *testing.T) {
	testCases := map[string]struct {
		container          *ResourceContainer
		expectedRoute      string
		expectedType       string
		expectedIdentifier string
	}{
		account: {
			container:          AccountIdentifier("abcd1234"),
			expectedRoute:      accounts,
			expectedType:       account,
			expectedIdentifier: "abcd1234",
		},
		zone: {
			container:          ZoneIdentifier("abcd1234"),
			expectedRoute:      zones,
			expectedType:       zone,
			expectedIdentifier: "abcd1234",
		},
		user: {
			container:          UserIdentifier("abcd1234"),
			expectedRoute:      user,
			expectedType:       user,
			expectedIdentifier: "abcd1234",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			setup()
			defer teardown()

			assert.Equal(t, tc.container.Level.String(), tc.expectedRoute)
			assert.Equal(t, tc.container.Type.String(), tc.expectedType)
			assert.Equal(t, tc.container.Identifier, tc.expectedIdentifier)
		})
	}
}
func TestResourcURLFragment(t *testing.T) {
	tests := map[string]struct {
		container *ResourceContainer
		want      string
	}{
		"account resource": {container: AccountIdentifier("foo"), want: "accounts/foo"},
		"zone resource":    {container: ZoneIdentifier("foo"), want: "zones/foo"},
		// this is pretty well deprecated in favour of `AccountIdentifier` but
		// here for completeness.
		"user level resource":    {container: UserIdentifier("foo"), want: "user"},
		"missing level resource": {container: &ResourceContainer{Level: "", Identifier: "foo"}, want: "foo"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := tc.container.URLFragment()
			assert.Equal(t, tc.want, got)
		})
	}
}
