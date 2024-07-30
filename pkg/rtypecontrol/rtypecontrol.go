package rtypecontrol

var validTypes = map[string]RegisterTypeOpts{}

type RegisterTypeOpts = struct {
	Name          string
	FromRawArgsFn func(items []any) (any, error)
}

func Register(ri RegisterTypeOpts) {

	// Does this already exist?
	if _, ok := validTypes[ri.Name]; ok {
		panic("rtype %q already registered. Can't register it a second time!")
	}

	validTypes[ri.Name] = ri

	// Do it the old way for backwards compatibility.
	RegisterCustomRecordType(ri.Name, "", "")
}

func IsValid(t string) bool {
	_, ok := validTypes[t]
	return ok
}

func Info(name string) RegisterTypeOpts {
	return validTypes[name]
}

// Legacy functions

// CustomRType stores an rtype that is only valid for this DSP.
type CustomRType struct {
	Name     string
	Provider string
	RealType string
}

// RegisterCustomRecordType registers a record type that is only valid for one provider.
// provider is the registered type of provider this is valid with
// name is the record type as it will appear in the js. (should be something like $PROVIDER_FOO)
// realType is the record type it will be replaced with after validation
func RegisterCustomRecordType(name, provider, realType string) {
	customRecordTypes[name] = &CustomRType{Name: name, Provider: provider, RealType: realType}
}

// GetCustomRecordType returns a registered custom record type, or nil if none
func GetCustomRecordType(rType string) *CustomRType {
	return customRecordTypes[rType]
}

var customRecordTypes = map[string]*CustomRType{}
