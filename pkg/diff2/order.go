package diff2

import "log"

type nameMap = map[string]struct{}

func orderByDependancies(changes ChangeList) ChangeList {
	resolvedNames := nameMap{}

	workingset := make(ChangeList, len(changes))
	newWorkingset := make(ChangeList, 0)

	copy(workingset, changes)

	var orderedChanges ChangeList

	for len(workingset) > 0 {
		added := 0

		for _, change := range workingset {
			if isDependancyFree(resolvedNames, change) {
				orderedChanges = append(orderedChanges, change)
				added += 1
			} else {
				newWorkingset = append(newWorkingset, change)
			}
		}

		if added == 0 && len(workingset) > 0 {
			log.Fatalf("The DNS changes appear to have unresolved dependancies like %v", workingset[0].HintDependancies)
			orderedChanges = append(orderedChanges, workingset...)
			break
		}

		workingset = newWorkingset
	}

	return orderedChanges
}

func isDependancyFree(resolvedNames nameMap, change Change) bool {
	for _, dependancy := range change.HintDependancies {
		if _, ok := resolvedNames[dependancy]; !ok {
			return false
		}
	}

	return true
}
