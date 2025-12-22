package none

import "github.com/StackExchange/dnscontrol/v4/pkg/providers"

/* The none provider does nothing. It can be used as a placeholder for third party providers or when you don't want the provider to perform any actions. */

// None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.
type none struct{}

func init() {
	providers.Register(
		providers.RegisterOpts{
			Name: "NONE",
			//NameAliases:        []string{},
			MaintainerGithubID: "@tlimoncelli",
			SupportLevel:       providers.SupportLevelOfficial,
			ProviderHandle:     none{},

			// Legacy functions:
			//RegistrarInitializer:          newReg,
			//DNSServiceProviderInitializer: newCloudflare,
			//RecordAuditor:                 AuditRecords,

			// Fields in the creds.json file:
			//CredsFields: []string{ },

			// Fields in the REGISTRAR("credkey", { metafield: "foo" })
			//MetadataFields: []string{},

			// DNS RecordTypes supported:
			//RecordTypes: []string{},

			Features: providers.DocumentationNotes{
				// The default for unlisted capabilities is 'Cannot'.
				// See providers/capabilities.go for the entire list of capabilities.
				providers.CanConcur:           providers.Can(),
				providers.CanUseDSForChildren: providers.Can(),
				providers.DocDualHost:         providers.Can(),
			},
		})
}
