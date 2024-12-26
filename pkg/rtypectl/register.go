package rtypectl

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"

	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
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

func Register(typeName string, opts RegisterOpts) error {

	printer.Printf("rtypectl.Register(%q)\n", typeName)

	// TODO(tlim): check for duplicates.

	if rtypeDB == nil {
		rtypeDB = map[string]RegisterOpts{}
	}

	if _, ok := rtypeDB[typeName]; ok {
		return fmt.Errorf("rtype %q already registered", typeName)
	}
	rtypeDB[typeName] = opts

	return nil
}
