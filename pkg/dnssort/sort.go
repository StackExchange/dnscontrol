package dnssort

import (
	"log"
)

type SortableRecord interface {
	GetNameFQDN() string
	GetFQDNDependencies() []string
}

type SortableRecords []SortableRecord

type unresolvedRecord struct {
	unresolvedDependencies nameMap
	record                 SortableRecord
}

func makeUnresolvedRecord(record SortableRecord) unresolvedRecord {
	unresolvedDependencies := nameMap{}

	for _, fqdn := range record.GetFQDNDependencies() {
		unresolvedDependencies.add(fqdn)
	}

	return unresolvedRecord{
		unresolvedDependencies: unresolvedDependencies,
		record:                 record,
	}
}

type DependencySortResult struct {
	SortedRecords     []SortableRecord
	UnresolvedRecords []SortableRecord
}

func SortByDependencies(records []SortableRecord) DependencySortResult {
	resolvedNames := CreateTree()

	workingset := make([]unresolvedRecord, len(records))
	var newWorkingset []unresolvedRecord

	for iX, record := range records {
		workingset[iX] = makeUnresolvedRecord(record)
	}

	var sortedRecords []SortableRecord
	var unresolvedRecords []SortableRecord

	for len(workingset) > 0 {
		added := 0

		for _, unresolved := range workingset {
			unresolved.unresolvedDependencies.removeAll(resolvedNames)

			if !unresolved.unresolvedDependencies.empty() {
				newWorkingset = append(newWorkingset, unresolved)

			} else {
				sortedRecords = append(sortedRecords, unresolved.record)
				resolvedNames.AddFQDN(unresolved.record.GetNameFQDN())
				added += 1
			}
		}

		if added == 0 && len(workingset) > 0 {
			log.Printf("The DNS changes appear to have unresolved dependencies like %s\n", workingset[0].unresolvedDependencies)
			for _, unresolved := range workingset {
				sortedRecords = append(sortedRecords, unresolved.record)
				unresolvedRecords = append(unresolvedRecords, unresolved.record)
			}
			break
		}

		workingset = newWorkingset
		newWorkingset = nil
	}

	return DependencySortResult{
		SortedRecords:     sortedRecords,
		UnresolvedRecords: unresolvedRecords,
	}
}
