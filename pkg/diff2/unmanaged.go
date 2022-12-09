package diff2

import (
	"strings"

	"github.com/gobwas/glob"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
)

func handsoff(
	existing, desired models.Records,
	unmanaged []*models.UnmanagedConfig) (models.Records, error) {

	// What foreign items should we ignore?
	foreign, err := manyQueries(existing, unmanaged)
	if err != nil {
		return nil, err
	}
	if len(foreign) != 0 {
		printer.Printf("INFO: Foreign records being ignored: (%d)\n", len(foreign))
		for i, r := range foreign {
			printer.Printf("- % 4d: %s %s %s\n", i, r.GetLabelFQDN(), r.Type, r.GetTargetRFC1035Quoted())
		}
	}

	// What desired items might conflict?
	conflicts, err := manyQueries(desired, unmanaged)
	if err != nil {
		return nil, err
	}
	if len(conflicts) != 0 {
		printer.Printf("WARN: dnsconfig.js records that overlap MANAGED: (%d)\n", len(conflicts))
		for i, r := range conflicts {
			printer.Printf("- % 4d: %s %s %s\n", i, r.GetLabelFQDN(), r.Type, r.GetTargetRFC1035Quoted())
		}
	}

	// Add the foreign items to the desired list.
	// (Rather than literally ignoring them, we just add them to the desired list
	// and all the diffing algorithms become more simple.)
	desired = append(desired, foreign...)

	return desired, nil
}

func manyQueries(rcs models.Records, queries []*models.UnmanagedConfig) (result models.Records, err error) {

	for _, q := range queries {

		lab := q.Label
		if lab == "" {
			lab = "*"
		}
		glabel, err := glob.Compile(lab)
		if err != nil {
			return nil, err
		}

		targ := q.Target
		if targ == "" {
			targ = "*"
		}
		gtarget, err := glob.Compile(targ)
		if err != nil {
			return nil, err
		}

		hasRType := compileTypeGlob(q.RType)

		for _, rc := range rcs {
			if match(rc, glabel, gtarget, hasRType) {
				result = append(result, rc)
			}
		}
	}
	return result, nil
}

func compileTypeGlob(g string) map[string]bool {
	m := map[string]bool{}
	for _, j := range strings.Split(g, ",") {
		m[strings.TrimSpace(j)] = true
	}
	return m
}

func match(rc *models.RecordConfig, glabel, gtarget glob.Glob, hasRType map[string]bool) bool {
	printer.Printf("DEBUG: match(%v, %v, %v, %v)\n", rc.NameFQDN, glabel, gtarget, hasRType)

	// _ = glabel.Match(rc.NameFQDN)
	// _ = matchType(rc.Type, hasRType)
	// x := rc.GetTargetField()
	// _ = gtarget.Match(x)

	if !glabel.Match(rc.NameFQDN) {
		printer.Printf("DEBUG: REJECTED LABEL: %s:%v\n", rc.NameFQDN, glabel)
		return false
	} else if !matchType(rc.Type, hasRType) {
		printer.Printf("DEBUG: REJECTED TYPE: %s:%v\n", rc.Type, hasRType)
		return false
	} else if gtarget == nil {
		return true
	} else if !gtarget.Match(rc.GetTargetField()) {
		printer.Printf("DEBUG: REJECTED TARGET: %v:%v\n", rc.GetTargetField(), gtarget)
		return false
	}
	return true
}

func matchType(s string, hasRType map[string]bool) bool {
	printer.Printf("DEBUG: matchType map=%v\n", hasRType)
	if len(hasRType) == 0 {
		return true
	}
	_, ok := hasRType[s]
	return ok
}
