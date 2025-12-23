package none

import (
	"encoding/json"

	"github.com/StackExchange/dnscontrol/v4/pkg/providers"
)

/* The none provider does nothing. It can be used as a placeholder for third party providers or when you don't want the provider to perform any actions. */

// None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.
type none struct{}

func init() {
	providers.Register(
		providers.RegisterOpts{
			Name:               "NONE",
			MaintainerGithubID: "@TomOnTime",
			SupportLevel:       providers.SupportLevelOfficial,
			ProviderHandle:     &none{},

			RegistrarInitializer:          newNoneReg,
			DNSServiceProviderInitializer: newNoneDSP,
			RecordAuditor:                 AuditRecords,

			Features: providers.DocumentationNotes{
				// The default for unlisted capabilities is 'Cannot'.
				// See providers/capabilities.go for the entire list of capabilities.
				providers.CanConcur:           providers.Can(),
				providers.CanUseDSForChildren: providers.Can(),
				providers.DocDualHost:         providers.Can(),
			},
		})
}

func newNoneDSP(_ map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	return &none{}, nil
}

func newNoneReg(_ map[string]string) (providers.Registrar, error) {
	return &none{}, nil
}
