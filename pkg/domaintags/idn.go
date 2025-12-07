package domaintags

import "golang.org/x/net/idna"

func EfficientToASCII(name string) string {
	nameIDN, err := idna.ToASCII(name)
	if err != nil {
		return name // Fallback to raw name on error.
	} else {
		// Avoid pointless duplication.
		if nameIDN == name {
			return name
		}
	}
	return nameIDN
}

func EfficientToUnicode(name string) string {
	nameUnicode, err := idna.ToUnicode(name)
	if err != nil {
		return name // Fallback to raw name on error.
	} else {
		// Avoid pointless duplication.
		if nameUnicode == name {
			return name
		}
	}
	return nameUnicode
}
