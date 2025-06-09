package none

import "github.com/StackExchange/dnscontrol/v4/providers"

/*
Provider: None

None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.

*/

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
