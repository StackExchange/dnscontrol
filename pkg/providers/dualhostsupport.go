package providers

//go:generate stringer -type DualHostSupport -trimprefix DualHostSupport

// DualHostSupport indicates the provider's level of support.
//
//	if x <  DualHostSupportSupported { // not supported }
//	if x >= DualHostSupportSupported { // supported }
type DualHostSupport int

const (
	// Enums that mean "not supported"
	DualHostSupportNotSupported DualHostSupport = iota
	DualHostSupportUntested
	DualHostSupportUnimplemented // Provider supports this but DNSControl does not have code to support it

	// Enums that mean "supported"
	DualHostSupportAdditionsOnly  // Provider does not permit modifying/deleting provider's NS records, only adding/removing additional records
	DualHostSupportItsComplicated // Provider supports this, but check the documentation for limits and oddities.
	DualHostSupportFull

	// Pivot point. Below here means "not supported". Above here means "supported".
	DualHostSupportSupported = DualHostSupportAdditionsOnly
)
