package rwth

import (
	"encoding/json"
	"errors"

	"github.com/DNSControl/dnscontrol/v5/pkg/providers"
)

type rwthProvider struct {
	apiToken string
	zones    map[string]zone
}

func init() {
	providers.Register[*rwthProvider]("RWTH", providers.Definition{
		FriendlyName: "RWTH Aachen",
		Maintainer:   "@mistererwin",
		SupportedTypes: []string{
			"Basic8",
			"PTR:Can:PTR records with empty targets are not supported",
			"SSHFP",
			"DS:Unimplemented:DS records are only supported at the apex and require a different API call that hasn't been implemented yet.",
			"SRV:Can:SRV records with empty targets are not supported.",
		},
		CanAutoDNSSEC:          providers.Unimplemented("Supported by RWTH but not implemented yet."),
		CanConcur:              providers.Unimplemented(),
		DocDualHost:            providers.Cannot(),
		DocOfficiallySupported: providers.Cannot(),
	})
}

// Initialize initializes a fresh provider instance.
func (api *rwthProvider) Initialize(settings map[string]string, _ json.RawMessage, _ *providers.CreateOptions) error {
	if settings["api_token"] == "" {
		return errors.New("missing RWTH api_token")
	}

	*api = rwthProvider{apiToken: settings["api_token"]}

	return nil
}
