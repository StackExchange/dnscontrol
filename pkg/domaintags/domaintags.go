package domaintags

import (
	"strings"

	"golang.org/x/net/idna"
)

// DomainFixedForms stores the various fixed forms of a domain name and tag.
type DomainFixedForms struct {
	NameRaw     string // "originalinput.com" (name as input by the user, lowercased (no tag))
	NameASCII   string // "punycode.com"
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
	var tag, nameRaw, nameASCII, nameUnicode, uniqueName string
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

	nameRaw = p[0]
	if strings.HasPrefix(n, nameRaw) {
		// Avoid pointless duplication.
		nameRaw = n[0:len(nameRaw)]
	}

	nameASCII, err = idna.ToASCII(nameRaw)
	if err != nil {
		nameASCII = nameRaw // Fallback to raw name on error.
	} else {
		nameASCII = strings.ToLower(nameASCII)
		// Avoid pointless duplication.
		if strings.HasPrefix(n, nameASCII) {
			// Avoid pointless duplication.
			nameASCII = n[0:len(nameASCII)]
		}
	}

	nameUnicode, err = idna.ToUnicode(nameASCII) // We use nameASCII since it is already lowercased.
	if err != nil {
		nameUnicode = nameRaw // Fallback to raw name on error.
	} else {
		// Avoid pointless duplication.
		if strings.HasPrefix(n, nameUnicode) {
			// Avoid pointless duplication.
			nameUnicode = n[0:len(nameUnicode)]
		}
	}

	if hasBang {
		uniqueName = nameASCII + "!" + tag
	} else {
		uniqueName = nameASCII
	}

	return DomainFixedForms{
		Tag:         tag,
		NameRaw:     nameRaw,
		NameASCII:   nameASCII,
		NameUnicode: nameUnicode,
		UniqueName:  uniqueName,
		HasBang:     hasBang,
	}
}
