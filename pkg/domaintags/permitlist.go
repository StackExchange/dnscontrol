package domaintags

import "strings"

type PermitList struct {
	// If the permit list is "all" or "".
	all   bool
	items []permitListItem
}
type permitListItem struct {
	tag, nameRaw, nameIDN, nameUnicode, uniqueName string
}

// CompilePermitList compiles a list of domain strings into a PermitList structure. The
func CompilePermitList(s string) PermitList {
	s = strings.TrimSpace(s)
	if s == "" || strings.ToLower(s) == "all" {
		return PermitList{all: true}
	}

	sl := PermitList{}
	for _, l := range strings.Split(s, ",") {
		l = strings.TrimSpace(l)
		if l == "" { // Skip empty entries. They match nothing.
			continue
		}
		tag, nameRaw, nameIDN, nameUnicode, uniqueName := MakeDomainFixForms(l)
		if tag == "" { // Treat empty tag as wildcard.
			tag = "*"
		}
		if nameIDN == "" { // Treat empty name as wildcard.
			nameIDN = "*"
		}
		sl.items = append(sl.items, permitListItem{
			tag:         tag,
			nameRaw:     nameRaw,
			nameIDN:     nameIDN,
			nameUnicode: nameUnicode,
			uniqueName:  uniqueName,
		})
	}

	return sl
}

func (pl *PermitList) Permitted(u string) bool {
	// If the permit list is "all", everything is permitted.
	if pl.all {
		return true
	}

	tag, _, nameIDN, nameUnicode, _ := MakeDomainFixForms(u)

	for _, item := range pl.items {
		// Skip if the tag doesn't match
		if item.tag != "*" && tag != item.tag {
			continue
		}
		// Now that we know the tag matches, we can focus on the name.

		if item.nameIDN == "*" {
			// `*!tag` or `*` matches everything.
			return true
		}
		// If the name starts with "*." then match the suffix.
		if strings.HasPrefix(item.nameIDN, "*.") {
			// example.com matches *.example.com
			if nameIDN == item.nameIDN[2:] || nameUnicode == item.nameUnicode[2:] {
				return true
			}
			// foo.example.com matches *.example.com
			if strings.HasSuffix(nameIDN, item.nameIDN[1:]) || strings.HasSuffix(nameUnicode, item.nameUnicode[1:]) {
				return true
			}
		}

		// No wildcards? Exact match.
		if item.nameIDN == nameIDN || item.nameUnicode == nameUnicode {
			return true
		}
	}

	return false
}
