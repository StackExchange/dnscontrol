package domaintags

import (
	"strings"

	"golang.org/x/net/idna"
)

// DomainFixedForms stores the various fixed forms of a domain name and tag.
type DomainFixedForms struct {
	NameRaw     string // "originalinput.com" (name as input by the user, lowercased (no tag))
	NameIDN     string // "punycode.com"
	NameUnicode string // "unicode.com" (converted to downcase BEFORE unicode conversion)
	UniqueName  string // "punycode.com!tag"

	Tag     string // The tag portion of `example.com!tag`
	HasBang bool   // Was there a "!" in the input when creating this struct?
}

// MakeDomainFixedForms turns the user-supplied name into the fixed forms.
// * .Tag: the domain tag (of "example.com!tag")
// * .NameRaw: lowercase version of how the user input the name in dnsconfig.js.
// * .Name: punycode version, downcased.
// * .NameUnicode: unicode version of the name, downcased.
// * .UniqueName: "example.com!tag" unique across the entire config.
func MakeDomainFixForms(n string) DomainFixedForms {
	var err error
	var tag, nameRaw, nameIDN, nameUnicode, uniqueName string
	var hasBang bool

	// Split tag from name.
	p := strings.SplitN(n, "!", 2)
	if len(p) == 2 {
		tag = p[1]
		hasBang = true
	} else {
		tag = ""
		hasBang = false
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

	if hasBang {
		uniqueName = nameIDN + "!" + tag
	} else {
		uniqueName = nameIDN
	}

	return DomainFixedForms{
		Tag:         tag,
		NameRaw:     nameRaw,
		NameIDN:     nameIDN,
		NameUnicode: nameUnicode,
		UniqueName:  uniqueName,
		HasBang:     hasBang,
	}
}
