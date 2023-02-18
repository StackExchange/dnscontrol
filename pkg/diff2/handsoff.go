package diff2

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/gobwas/glob"
)

/*

# How do NO_PURGE, IGNORE_*, ENSURE_ABSENT and friends work?


## Terminology:

* "existing" refers to the records downloaded from the provider via the API.
* "desired" refers to the records generated from dnsconfig.js.
* "absences" refers to a list of records tagged with ASSURE_ABSENT.

## What are the features?

There are 2 ways to tell DNSControl not to touch existing records in a domain,
and 1 way to make exceptions.

* NO_PURGE: Tells DNSControl not to delete records in a domain.
	* New records can be created; existing records (matched on label:rtype) can
	    be modified.
	* FYI: This means you can't have a label with two A records, one controlled
	    by DNSControl and one controlled by an external system.
* UNMANAGED( labelglob, typelist, targetglob):
    * "If an existing record matches this pattern, don't touch it!""
    * IGNORE_NAME(foo) is the same as UNMANAGED(foo, "*", "*")
    * IGNORE_TARGET(foo) is the same as UNMANAGED("*", "*", foo)
    * FYI: You CAN have a label with two A records, one controlled by
	    DNSControl and one controlled by an external system.  DNSControl would
		need to have an UNMANAGED() statement with a targetglob that matches
		the external system's target values.
* ASSURE_ABSENT: Override NO_PURGE for specific records. i.e. delete them even
    though NO_PURGE is enabled.
    * If any of these records are in desired (matched on
        label:rtype:target), remove them.  This takes priority over
		all of the above.

## Implementation premise

The fundamental premise is "if you don't want it deleted, put it in the
'desired' list." So, for example, if you want to IGNORE_NAME("www"), then you
find any records with the label "www" in "existing" and copy them to "desired".
As a result, the diff2 algorithm won't delete them because they are desired!

This is different than in the old system (pkg/diff) which would generate the
diff but but then do a bunch of checking to see if the record was one that
shouldn't be deleted.  Or, in the case of NO_PURGE, would simply not do the
deletions.  This was complex because there were many edge cases to deal with.
It was often also wrong. For example, if a provider updates all records in a
RecordSet at once, you shouldn't NOT update the record.

## Implementation

Here is how we intend to implement these features:

  UNMANAGED is implemented as:
  * Take the list of existing records. If any match one of the UNMANAGED glob
      patterns, add it to the "ignored list".
  * If any item on the "ignored list" is also in "desired" (match on
      label:rtype), output a warning (defeault) or declare an error (if
      DISABLE_UNMANAGED_SAFETY_CHECK is true).
  * Add the "ignore list" records to desired.

  NO_PURGE + ENSURE_ABSENT is implemented as:
  * Take the list of existing records. If any do not appear in desired, add them
      to desired UNLESS they appear in absences.
  * "appear in desired" is done by matching on label:type.
  * "appear in absences" is done by matching on label:type:target.

  However the actual changes are implemented as:
    foreach rec in existing:
	    if rec matches_any_unmanaged_pattern:
	        if rec in desired:
		    	if "DISABLE_UNMANAGED_SAFETY_CHECK" is false:
					Display a warning.
		     	else
			   		Return an error.
		  	add rec to "foreign list"
	    else:
	     	if NO_PURGE:
		 		if rec NOT in desired: (matched on label:type)
		 		    if rec NOT in absences: (matched on label:type:combinedtarget)
		      			Add rec to "foreign list"
	Append "foreign list" to "desired"

*/

func handsoff(
	domain string,
	existing, desired, absences models.Records,
	unmanagedConfigs []*models.UnmanagedConfig,
	unmanagedSafely bool,
	noPurge bool,
) (models.Records, error) {

	err := compileUnmanagedConfigs(unmanagedConfigs)
	if err != nil {
		return nil, err
	}

	ignorable, foreign := ignoreOrNoPurge(domain, existing, desired, absences, unmanagedConfigs, noPurge)
	if len(foreign) != 0 {
		printer.Printf("INFO: %d records not being deleted because of NO_PURGE:\n", len(foreign))
		for _, r := range foreign {
			printer.Printf("    %s. %s %s\n", r.GetLabelFQDN(), r.Type, r.GetTargetRFC1035Quoted())
		}
	}
	if len(ignorable) != 0 {
		printer.Printf("INFO: %d records not being deleted because of IGNORE*():\n", len(ignorable))
		for _, r := range ignorable {
			printer.Printf("    %s %s %s\n", r.GetLabelFQDN(), r.Type, r.GetTargetRFC1035Quoted())
		}
	}

	conflicts := findConflicts(unmanagedConfigs, desired)
	if len(conflicts) != 0 {
		printer.Printf("INFO: %d records that are both IGNORE*()'d and not ignored:\n", len(conflicts))
		for _, r := range conflicts {
			printer.Printf("    %s %s %s\n", r.GetLabelFQDN(), r.Type, r.GetTargetRFC1035Quoted())
		}
		if unmanagedSafely {
			return nil, fmt.Errorf("ERROR: Unsafe to continue. Add DISABLE_UNMANAGED_SAFETY_CHECK to D() to override")
		}
	}

	// Add the ignored/foreign items to the desired list so they are not deleted:
	desired = append(desired, ignorable...)
	desired = append(desired, foreign...)
	return desired, nil
}

func ignoreOrNoPurge(domain string, existing, desired, absences models.Records, unmanagedConfigs []*models.UnmanagedConfig, noPurge bool) (models.Records, models.Records) {
	var ignorable, foreign models.Records
	desiredDB := models.NewRecordDBFromRecords(desired, domain)
	//fmt.Printf("DEBUG: start absent\n")
	absentDB := models.NewRecordDBFromRecords(absences, domain)
	compileUnmanagedConfigs(unmanagedConfigs)
	//fmt.Printf("DEBUG: setup done\n")
	for _, rec := range existing {
		if matchAny(unmanagedConfigs, rec) {
			//fmt.Printf("DEBUG: IGNORING: %v %v\n", rec.GetLabelFQDN(), rec)
			ignorable = append(ignorable, rec)
		} else {
			if noPurge {
				// Is this a candidate for purging?
				if !desiredDB.ContainsLT(rec) {
					// Yes, but not if it is an exception!
					if !absentDB.ContainsLT(rec) {
						//fmt.Printf("DEBUG: FOREIGN: %v %v\n", rec.GetLabelFQDN(), rec)
						foreign = append(foreign, rec)
					}
				}
			}
		}
	}
	return ignorable, foreign
}

func findConflicts(uconfigs []*models.UnmanagedConfig, recs models.Records) models.Records {
	var conflicts models.Records
	for _, rec := range recs {
		if matchAny(uconfigs, rec) {
			conflicts = append(conflicts, rec)
		}
	}
	return conflicts
}

func compileUnmanagedConfigs(configs []*models.UnmanagedConfig) error {
	//fmt.Printf("DEBUG: compileUnmanagedConfigs(%v)\n", configs)
	var err error

	for i := range configs {
		c := configs[i]

		if c.LabelPattern == "" {
			c.LabelPattern = "*"
		}
		//fmt.Printf("DEBUG: compiling labelPattern: %q\n", c.LabelPattern)
		c.LabelGlob, err = glob.Compile(c.LabelPattern)
		if err != nil {
			return err
		}

		//fmt.Printf("DEBUG: compiling type: %q\n", c.RTypePattern)
		c.RTypeMap = make(map[string]struct{})
		if c.RTypePattern != "*" && c.RTypePattern != "" {
			for _, part := range strings.Split(c.RTypePattern, ",") {
				part = strings.TrimSpace(part)
				//fmt.Printf("    DEBUG: part=%q\n", part)
				c.RTypeMap[part] = struct{}{}
			}
		}
		//fmt.Printf("DEBUG: compiling type DONE\n")

		if c.TargetPattern == "" {
			c.TargetPattern = "*"
		}
		//fmt.Printf("DEBUG: compiling targetPattern: %q\n", c.TargetPattern)
		c.TargetGlob, err = glob.Compile(c.TargetPattern)
		if err != nil {
			return err
		}
	}
	return nil
}

func matchAny(uconfigs []*models.UnmanagedConfig, rec *models.RecordConfig) bool {
	//fmt.Printf("DEBUG: matchAny: rec: %v %v\n", rec.GetLabel(), rec)
	//for _, j := range uconfigs {
	//	fmt.Printf("    DEBUG: uconfig = %+v\n", j)
	//}
	for _, uc := range uconfigs {
		if matchLabel(uc.LabelGlob, rec.GetLabel()) &&
			matchType(uc.RTypeMap, rec.Type) &&
			matchTarget(uc.TargetGlob, rec.GetLabel()) {
			//fmt.Printf(" true\n")
			return true
		}
	}
	//fmt.Printf(" false\n")
	return false
}
func matchLabel(labelGlob glob.Glob, labelName string) bool {
	if labelGlob == nil {
		return true
	}
	return labelGlob.Match(labelName)
}
func matchType(typeMap map[string]struct{}, typeName string) bool {
	if len(typeMap) == 0 {
		return true
	}
	_, ok := typeMap[typeName]
	return ok
}
func matchTarget(targetGlob glob.Glob, targetName string) bool {
	if targetGlob == nil {
		return true
	}
	return targetGlob.Match(targetName)
}

// func manyQueries(rcs models.Records, queries []*models.UnmanagedConfig) (result models.Records, err error) {

// 	for _, q := range queries {

// 		lab := q.Label
// 		if lab == "" {
// 			lab = "*"
// 		}
// 		glabel, err := glob.Compile(lab)
// 		if err != nil {
// 			return nil, err
// 		}

// 		targ := q.Target
// 		if targ == "" {
// 			targ = "*"
// 		}
// 		gtarget, err := glob.Compile(targ)
// 		if err != nil {
// 			return nil, err
// 		}

// 		hasRType := compileTypeGlob(q.RType)

// 		for _, rc := range rcs {
// 			if match(rc, glabel, gtarget, hasRType) {
// 				result = append(result, rc)
// 			}
// 		}
// 	}
// 	return result, nil
// }

// func compileTypeGlob(g string) map[string]bool {
// 	m := map[string]bool{}
// 	for _, j := range strings.Split(g, ",") {
// 		m[strings.TrimSpace(j)] = true
// 	}
// 	return m
// }

// func match(rc *models.RecordConfig, glabel, gtarget glob.Glob, hasRType map[string]bool) bool {
// 	//printer.Printf("DEBUG: match(%v, %v, %v, %v)\n", rc.NameFQDN, glabel, gtarget, hasRType)

// 	// _ = glabel.Match(rc.NameFQDN)
// 	// _ = matchType(rc.Type, hasRType)
// 	// x := rc.GetTargetField()
// 	// _ = gtarget.Match(x)

// 	if !glabel.Match(rc.NameFQDN) {
// 		//printer.Printf("DEBUG: REJECTED LABEL: %s:%v\n", rc.NameFQDN, glabel)
// 		return false
// 	} else if !matchType(rc.Type, hasRType) {
// 		//printer.Printf("DEBUG: REJECTED TYPE: %s:%v\n", rc.Type, hasRType)
// 		return false
// 	} else if gtarget == nil {
// 		return true
// 	} else if !gtarget.Match(rc.GetTargetField()) {
// 		//printer.Printf("DEBUG: REJECTED TARGET: %v:%v\n", rc.GetTargetField(), gtarget)
// 		return false
// 	}
// 	return true
// }

// func matchType(s string, hasRType map[string]bool) bool {
// 	//printer.Printf("DEBUG: matchType map=%v\n", hasRType)
// 	if len(hasRType) == 0 {
// 		return true
// 	}
// 	_, ok := hasRType[s]
// 	return ok
// }
