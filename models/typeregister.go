package models

import (
	"fmt"
)

type FromRawFn func(rc *RecordConfig, rawfields []string, metadata map[string]string, origin string) error

type RegisterOpts struct {
	Enum            int
	PopulateFromRaw FromRawFn
}

var rtypeDB map[string]RegisterOpts

var validTypes = map[string]struct{}{}

// MustRegisterType registers a new record type with the system. Use it in init() functions.
func MustRegisterType(typeName string, opts RegisterOpts) {

	if rtypeDB == nil {
		rtypeDB = map[string]RegisterOpts{}
	}

	if _, ok := rtypeDB[typeName]; ok {
		panic(fmt.Errorf("rtype %q already registered", typeName))
	}
	rtypeDB[typeName] = opts

	// Legacy version:

	// Does this already exist?
	if _, ok := validTypes[typeName]; ok {
		panic("rtype %q already registered. Can't register it a second time!")
	}
	validTypes[typeName] = struct{}{}
}

// GetTypeOps returns the RegisterOpts for a given record type.
func GetTypeOps(t string) (*RegisterOpts, error) {
	if opts, ok := rtypeDB[t]; ok {
		return &opts, nil
	}
	return nil, fmt.Errorf("rtype %q not found", t)
}

// IsValid returns true if the string t is a valid type ("valid" means that it has been registered).
func IsValid(t string) bool {
	_, ok := validTypes[t]
	return ok
}

// IsTypeLegacy returns true if the type has NOT been converted to the new way of doing types.
func IsTypeLegacy(t string) bool {
	_, ok := validTypes[t]
	return !ok
}

// IsTypeUpgraded returns true if the type has been converted to the new way of doing types.
func IsTypeUpgraded(t string) bool {
	_, ok := validTypes[t]
	return ok
}
