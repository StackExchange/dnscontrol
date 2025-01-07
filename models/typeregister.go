package models

import (
	"fmt"
)

type FromRawFn func(rc *RecordConfig, rawfields []string, metadata map[string]string, origin string) error

type RegisterOpts struct {
	Enum    int
	FromRaw FromRawFn
}

var rtypeDB map[string]RegisterOpts

var validTypes = map[string]struct{}{}

func RegisterType(typeName string, opts RegisterOpts) error {

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
	//providers.RegisterCustomRecordType(typeName, "CLOUDFLAREAPI", "")

	return nil
}

func GetTypeOps(t string) (*RegisterOpts, error) {
	if opts, ok := rtypeDB[t]; ok {
		return &opts, nil
	}
	return nil, fmt.Errorf("rtype %q not found", t)
}

func IsValid(t string) bool {
	_, ok := validTypes[t]
	return ok
}

func IsTypeLegacy(t string) bool {
	_, ok := validTypes[t]
	return !ok
}

func IsTypeUpgraded(t string) bool {
	_, ok := validTypes[t]
	return ok
}
