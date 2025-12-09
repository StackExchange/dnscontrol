package domaintags

import "golang.org/x/net/idna"

// EfficientToASCII converts a domain name to its ASCII representation using IDNA, on error returns the original name, and avoids wasting memory when possible.
func EfficientToASCII(name string) string {
	nameIDN, err := idna.ToASCII(name)
	if err != nil {
		return name // Fallback to raw name on error.
	}
	// Avoid pointless duplication.
	if nameIDN == name {
		return name
	}
	return nameIDN
}

// EfficientToUnicode converts a domain name to its Unicode representation using IDNA, on error returns the original name, and avoids wasting memory when possible.
func EfficientToUnicode(name string) string {
	nameUnicode, err := idna.ToUnicode(name)
	if err != nil {
		return name // Fallback to raw name on error.
	}
	// Avoid pointless duplication.
	if nameUnicode == name {
		return name
	}
	return nameUnicode
}
