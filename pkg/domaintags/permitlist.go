package domaintags

import (
	"strings"
)

type PermitList struct {
	// If the permit list is "all" or "".
	all   bool
	items []DomainFixedForms
}

// CompilePermitList compiles a list of domain strings into a PermitList structure. The
func CompilePermitList(s string) PermitList {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" || strings.ToLower(s) == "all" {
		return PermitList{all: true}
	}

	sl := PermitList{}
	for _, l := range strings.Split(s, ",") {
		l = strings.TrimSpace(l)
		if l == "" { // Skip empty entries. They match nothing.
			continue
		}
		ff := MakeDomainFixForms(l)
		if ff.HasBang && ff.NameASCII == "" { // Treat empty name as wildcard.
			ff.NameASCII = "*"
		}
		sl.items = append(sl.items, ff)
	}

	return sl
}

func (pl *PermitList) Permitted(domToCheck string) bool {

	// If the permit list is "all", everything is permitted.
	if pl.all {
		return true
	}

	domToCheckFF := MakeDomainFixForms(domToCheck)

	for _, filterItem := range pl.items {

		// Special case: filter=example.com!* does not match example.com (no tag)
		if filterItem.Tag == "*" && !domToCheckFF.HasBang {
			continue
		}
		// Special case: filter=example.com!* does not match example.com! (empty tag)
		if filterItem.Tag == "*" && domToCheckFF.HasBang && domToCheckFF.Tag == "" {
			continue
		}
		// Special case: filter=example.com! does not match example.com!tag
		if filterItem.HasBang && filterItem.Tag == "" && domToCheckFF.HasBang && domToCheckFF.Tag != "" {
			continue
		}

		// Skip if tags don't match
		if (filterItem.Tag != "*") && (domToCheckFF.Tag != filterItem.Tag) {
			continue
		}

		// Now that we know the tag matches, we can focus on the name.

		// `*!tag` or `*` matches everything.
		if filterItem.NameASCII == "*" {
			return true
		}

		// If the name starts with "*." then match the suffix.
		if strings.HasPrefix(filterItem.NameASCII, "*.") {
			// example.com matches *.example.com
			if domToCheckFF.NameASCII == filterItem.NameASCII[2:] || domToCheckFF.NameUnicode == filterItem.NameUnicode[2:] {
				return true
			}
			// foo.example.com matches *.example.com
			if strings.HasSuffix(domToCheckFF.NameASCII, filterItem.NameASCII[1:]) || strings.HasSuffix(domToCheckFF.NameUnicode, filterItem.NameUnicode[1:]) {
				return true
			}
		}

		// No wildcards? Exact match.
		if filterItem.NameASCII == domToCheckFF.NameASCII || filterItem.NameUnicode == domToCheckFF.NameUnicode {
			return true
		}
	}

	return false
}
