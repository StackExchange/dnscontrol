package rtypecontrol

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// backwards compatibility:
var validTypes = map[string]struct{}{}

type RType interface {
	// Returns the name of the rtype ("A", "MX", etc.)
	Name() string

	// RecordConfig factory. Updates a RecordConfig's fields based on args.
	FromArgs(*models.RecordConfig, args []any) (*models.RecordConfig, error) //

	// Returns a string representation of the record in RFC1038 format.
	// AsRFC1038String([]string) (string, error)
}

// Map of registered rtypes.
var Iface map[string]RType = map[string]RType{}

func Register(typeName string, t RType) {
	name := t.Name()
	if _, ok := Iface[name]; ok {
		panic(fmt.Sprintf("rtype %q already registered. Can't register it a second time!", name))
	}
	// Store the interface
	Iface[name] = t

	// For compatibility with legacy systems:
	providers.RegisterCustomRecordType(name, "", "")

}

func IsValid(name string) bool {
	_, ok := Iface[name]
	return ok
}
