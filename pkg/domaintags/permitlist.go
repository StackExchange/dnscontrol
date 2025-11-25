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
	//fmt.Printf("DEBUG: CompilePermitList(%q)\n", s)

	s = strings.TrimSpace(s)
	if s == "" || s == "*" || strings.ToLower(s) == "all" {
		//fmt.Printf("DEBUG: CompilePermitList: ALL\n")
		return PermitList{all: true}
	}

	sl := PermitList{}
	for _, l := range strings.Split(s, ",") {
		l = strings.TrimSpace(l)
		if l == "" { // Skip empty entries. They match nothing.
			continue
		}
		ff := MakeDomainFixForms(l)
		if ff.HasBang && ff.NameIDN == "" { // Treat empty name as wildcard.
			ff.NameIDN = "*"
		}
		sl.items = append(sl.items, ff)
	}

	//fmt.Printf("DEBUG: CompilePermitList: RETURN %+v\n", sl)
	return sl
}

func (pl *PermitList) Permitted(domToCheck string) bool {
	//fmt.Printf("DEBUG: Permitted(%q)\n", domToCheck)

	// If the permit list is "all", everything is permitted.
	if pl.all {
		//fmt.Printf("DEBUG: Permitted RETURN true\n")
		return true
	}

	domToCheckFF := MakeDomainFixForms(domToCheck)
	// fmt.Printf("DEBUG: input: %+v\n", domToCheckFF)

	for _, filterItem := range pl.items {
		// fmt.Printf("DEBUG: Checking item %+v\n", filterItem)

		// Special case: filter=example.com!* does not match example.com (no tag)
		if filterItem.Tag == "*" && !domToCheckFF.HasBang {
			// fmt.Printf("DEBUG: Skipping due to no tag present\n")
			continue
		}
		// Special case: filter=example.com!* does not match example.com! (empty tag)
		if filterItem.Tag == "*" && domToCheckFF.HasBang && domToCheckFF.Tag == "" {
			// fmt.Printf("DEBUG: Skipping due to empty tag present\n")
			continue
		}
		// Special case: filter=example.com! does not match example.com!tag
		if filterItem.HasBang && filterItem.Tag == "" && domToCheckFF.HasBang && domToCheckFF.Tag != "" {
			// fmt.Printf("DEBUG: Skipping due to non-empty tag present\n")
			continue
		}

		// Skip if the tag doesn't match
		if (filterItem.Tag != "*") && (domToCheckFF.Tag != filterItem.Tag) {
			continue
		}
		// Now that we know the tag matches, we can focus on the name.

		if filterItem.NameIDN == "*" {
			// `*!tag` or `*` matches everything.
			return true
		}
		// If the name starts with "*." then match the suffix.
		if strings.HasPrefix(filterItem.NameIDN, "*.") {
			// example.com matches *.example.com
			if domToCheckFF.NameIDN == filterItem.NameIDN[2:] || domToCheckFF.NameUnicode == filterItem.NameUnicode[2:] {
				return true
			}
			// foo.example.com matches *.example.com
			if strings.HasSuffix(domToCheckFF.NameIDN, filterItem.NameIDN[1:]) || strings.HasSuffix(domToCheckFF.NameUnicode, filterItem.NameUnicode[1:]) {
				return true
			}
		}

		// No wildcards? Exact match.
		if filterItem.NameIDN == domToCheckFF.NameIDN || filterItem.NameUnicode == domToCheckFF.NameUnicode {
			return true
		}
	}

	return false
}
