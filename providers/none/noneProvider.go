package none

import "github.com/StackExchange/dnscontrol/v4/providers"

/*
Provider: None

None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.

*/

// None is the struct that is used to hold state for the provider.
//
// This same struct is used for all domains that use this provider. Thus
// per-domain data should be stored as a map[domain]thing and be protected by
// mutexes.
type None struct{}

var featuresNone = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanConcur: providers.Can(),
}

func init() {
	providers.RegisterRegistrarType("NONE", func(map[string]string) (providers.Registrar, error) {
		return None{}, nil
	}, featuresNone)
}
