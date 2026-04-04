package domaintags

import (
	"strings"
)

// PermitList is a structure that holds a pre-compiled version of the --domains
// commmand line argument.  "all" means all domains are permitted and the rest
// of the list is ignored. Otherwise, the list contains each element stored in a
// variety of ways useful to the matching algorithm.
type PermitList struct {
	// If the permit list is "all" or "".
	all   bool
	items []*DomainNameVarieties
}

// CompilePermitList compiles a list of domain strings into a PermitList structure.
func CompilePermitList(s string) PermitList {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" || strings.ToLower(s) == "all" {
		return PermitList{all: true}
	}

	sl := PermitList{}
	for l := range strings.SplitSeq(s, ",") {
		l = strings.TrimSpace(l)
		if l == "" { // Skip empty entries. They match nothing.
			continue
		}
		ff := MakeDomainNameVarieties(l)
		if ff.HasBang && ff.NameASCII == "" { // Treat empty name as wildcard.
			ff.NameASCII = "*"
		}
		sl.items = append(sl.items, ff)
	}

	return sl
}

// Permitted returns whether a domain is permitted by the PermitList.
func (pl *PermitList) Permitted(domToCheck string) bool {

	// If the permit list is "all", everything is permitted.
	if pl.all {
		return true
	}

	domToCheckFF := MakeDomainNameVarieties(domToCheck)

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
