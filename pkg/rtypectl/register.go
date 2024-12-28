package rtypectl

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

type FromRawFn func(*models.RecordConfig, string, []any, map[string]string) error

// type FromRawResult struct {
// 	LabelShort, LabelFQDN string
// 	//LabelDisplayFQDN      string
// 	Fields     interface{}
// 	Comparable string
// 	Display    string
// }

type RegisterOpts struct {
	Enum    int
	FromRaw FromRawFn
}

var rtypeDB map[string]RegisterOpts

var validTypes = map[string]struct{}{}

func Register(typeName string, opts RegisterOpts) error {

	//printer.Printf("rtypectl.Register(%q)\n", typeName)

	if rtypeDB == nil {
		rtypeDB = map[string]RegisterOpts{}
	}

	if _, ok := rtypeDB[typeName]; ok {
		return fmt.Errorf("rtype %q already registered", typeName)
	}
	rtypeDB[typeName] = opts

	// Legacy version:

	// Does this already exist?
	if _, ok := validTypes[typeName]; ok {
		panic("rtype %q already registered. Can't register it a second time!")
	}
	validTypes[typeName] = struct{}{}
	providers.RegisterCustomRecordType(typeName, "CLOUDFLAREAPI", "")

	return nil
}

func IsValid(t string) bool {
	_, ok := validTypes[t]
	return ok
}
