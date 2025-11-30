package rtypecontrol

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// backwards compatibility:
//var validTypes = map[string]struct{}{}

type RType interface {
	// Returns the name of the rtype ("A", "MX", etc.)
	Name() string

	// RecordConfig factory. Updates a RecordConfig's fields based on args.
	FromArgs(*models.DomainConfig, *models.RecordConfig, []any) error

	CopyToLegacyFields(*models.RecordConfig)
}

// Map of registered rtypes.
var Func map[string]RType = map[string]RType{}

func Register(t RType) {
	name := t.Name()
	if _, ok := Func[name]; ok {
		panic(fmt.Sprintf("rtype %q already registered. Can't register it a second time!", name))
	}
	// Store the interface
	Func[name] = t

	// For compatibility with legacy systems:
	providers.RegisterCustomRecordType(name, "", "")
}

func IsModernType(name string) bool {
	_, ok := Func[name]
	return ok
}
