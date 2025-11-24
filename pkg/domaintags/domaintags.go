package domaintags

import (
	"strings"

	"golang.org/x/net/idna"
)

// MakeDomainFixedForms turns the user-supplied name into the fixed forms.
// * .Tag: the domain tag (of "example.com!tag")
// * .NameRaw: lowercase version of how the user input the name in dnsconfig.js.
// * .Name: punycode version, downcased.
// * .NameUnicode: unicode version of the name, downcased.
// * .UniqueName: "example.com!tag" unique across the entire config.
func MakeDomainFixForms(n string) (tag, nameRaw, nameIDN, nameUnicode, UniqueName string) {
	var err error

	p := strings.SplitN(n, "!", 2)
	if len(p) != 2 {
		tag = ""
	} else {
		tag = p[1]
	}

	nameRaw = strings.ToLower(p[0])
	if strings.HasPrefix(n, nameRaw) {
		// Avoid pointless duplication.
		nameRaw = n[0:len(nameRaw)]
	}

	nameIDN, err = idna.ToASCII(nameRaw)
	if err != nil {
		nameIDN = nameRaw // Fallback to raw name on error.
	} else {
		// Avoid pointless duplication.
		if nameIDN == nameRaw {
			nameIDN = nameRaw
		}
	}

	nameUnicode, err = idna.ToUnicode(nameRaw)
	if err != nil {
		nameUnicode = nameRaw // Fallback to raw name on error.
	} else {
		// Avoid pointless duplication.
		if nameUnicode == nameRaw {
			nameUnicode = nameRaw
		}
	}

	UniqueName = nameIDN + "!" + tag

	return
}
