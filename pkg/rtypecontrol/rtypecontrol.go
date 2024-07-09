package rtypecontrol

import "github.com/StackExchange/dnscontrol/v4/providers"

var validTypes = map[string]struct{}{}

func Register(t string) {
	// Does this already exist?
	if _, ok := validTypes[t]; ok {
		panic("rtype %q already registered. Can't register it a second time!")
	}

	validTypes[t] = struct{}{}

	providers.RegisterCustomRecordType(t, "", "")
}

func IsValid(t string) bool {
	_, ok := validTypes[t]
	return ok
}
